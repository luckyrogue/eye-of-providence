//! Local SQLite encryption helper.
//!
//! Зачем:
//!   README §3.1 / §8 обещают «шифрование с ключом из Keychain/DPAPI», но до
//!   этого модуля события писались в SQLite plaintext. Любой с file-system
//!   access к `~/Library/Application Support/eop/eop.sqlite` (или Windows
//!   `%APPDATA%\eop\eop.sqlite`) мог читать app bundles, project paths и пр.
//!
//! Что делаем:
//!   AES-256-GCM AEAD на каждую запись `event_buffer.payload`. Ключ хранится
//!   в системном secret store через тот же `keyring`-крейт, что и токен/
//!   backend_url (см. `core/auth.rs`).
//!
//! Запись на диске:
//!   `[12-byte nonce | ciphertext | 16-byte GCM auth tag]`
//!   Nonce — random per-row (rand::OsRng), не reused. Auth tag валидирует
//!   integrity: если кто-то flip-нул бит в SQLite, decrypt вернёт ошибку.
//!
//! Threat model:
//!   - **Защищает от**: пассивного file-read (другой user на машине,
//!     случайный backup, кража диска). Шифрование на rest.
//!   - **Не защищает от**: malware с теми же привилегиями (может прочитать
//!     ключ из Keychain через ту же `keyring` API). Это фундаментальное
//!     ограничение local-encryption — без TEE/Secure Enclave нельзя.
//!
//! Тесты:
//!   `LocalCrypto::with_key([u8;32])` — test-only factory, не трогает keyring.

use aes_gcm::aead::{Aead, KeyInit, OsRng};
use aes_gcm::{AeadCore, Aes256Gcm, Key, Nonce};
use anyhow::{anyhow, Context, Result};
use base64::engine::general_purpose::STANDARD as BASE64;
use base64::Engine;
use keyring::Entry;
use zeroize::Zeroize;

/// 12 байт — стандарт для AES-GCM nonce (96-bit IV). Не уменьшать: коллизии
/// nonce с тем же ключом ломают конфиденциальность катастрофически.
const NONCE_LEN: usize = 12;

/// Keyring service/account для DB ключа. Сервис совпадает с `core/auth.rs`
/// чтобы все секреты агента жили в одной «папке» Keychain/Credential Manager.
const KEYRING_SERVICE: &str = "eop-agent";
const KEYRING_ACCOUNT: &str = "db-encryption-key";

pub struct LocalCrypto {
    cipher: Aes256Gcm,
}

impl LocalCrypto {
    /// Открывает или создаёт DB-ключ в системном secret store.
    /// При первом запуске — генерим 32 байта rand::OsRng и сохраняем base64.
    /// При последующих — читаем существующий и инициализируем cipher.
    pub fn open_or_init() -> Result<Self> {
        let entry = Entry::new(KEYRING_SERVICE, KEYRING_ACCOUNT)
            .map_err(|e| anyhow!("keyring entry: {e}"))?;
        let key_bytes = match entry.get_password() {
            Ok(b64) => {
                let mut buf = BASE64
                    .decode(b64.as_bytes())
                    .context("decode base64 db key from keyring")?;
                if buf.len() != 32 {
                    // Corrupted/manually-edited — генерим заново и переписываем.
                    tracing::warn!(
                        actual_len = buf.len(),
                        "keyring db key has wrong length, regenerating"
                    );
                    buf.zeroize();
                    let fresh = generate_key();
                    let b64 = BASE64.encode(fresh);
                    entry
                        .set_password(&b64)
                        .context("rewrite db key into keyring")?;
                    fresh.to_vec()
                } else {
                    buf
                }
            }
            Err(keyring::Error::NoEntry) => {
                let fresh = generate_key();
                let b64 = BASE64.encode(fresh);
                entry
                    .set_password(&b64)
                    .context("persist new db key into keyring")?;
                tracing::info!("generated new local DB encryption key");
                fresh.to_vec()
            }
            Err(e) => return Err(anyhow!("keyring read db key: {e}")),
        };

        let cipher = Aes256Gcm::new(Key::<Aes256Gcm>::from_slice(&key_bytes));
        // Сразу обнуляем буфер ключа в памяти после конструирования cipher —
        // дальше работаем только через cipher API.
        let mut zero_buf = key_bytes;
        zero_buf.zeroize();
        Ok(Self { cipher })
    }

    /// Test-only: явный ключ, без обращения к keyring.
    #[cfg(test)]
    pub fn with_key(key: [u8; 32]) -> Self {
        let cipher = Aes256Gcm::new(Key::<Aes256Gcm>::from_slice(&key));
        Self { cipher }
    }

    /// Шифрует plaintext. Output: `[12B nonce][ciphertext+16B tag]`.
    pub fn seal(&self, plaintext: &[u8]) -> Result<Vec<u8>> {
        let nonce = Aes256Gcm::generate_nonce(&mut OsRng);
        let ct = self
            .cipher
            .encrypt(&nonce, plaintext)
            .map_err(|e| anyhow!("aes-gcm encrypt: {e}"))?;
        // [nonce | ct+tag] — `encrypt` уже включил auth tag в конец ct.
        let mut out = Vec::with_capacity(NONCE_LEN + ct.len());
        out.extend_from_slice(&nonce);
        out.extend_from_slice(&ct);
        Ok(out)
    }

    /// Расшифровывает запись. Возвращает ошибку при tag mismatch (data
    /// tampering / неверный ключ / битый файл).
    pub fn open(&self, record: &[u8]) -> Result<Vec<u8>> {
        if record.len() < NONCE_LEN + 16 {
            return Err(anyhow!(
                "encrypted record too short: {} bytes (min {})",
                record.len(),
                NONCE_LEN + 16
            ));
        }
        let (nonce_bytes, ct) = record.split_at(NONCE_LEN);
        let nonce = Nonce::from_slice(nonce_bytes);
        let pt = self
            .cipher
            .decrypt(nonce, ct)
            .map_err(|e| anyhow!("aes-gcm decrypt (tag mismatch / wrong key?): {e}"))?;
        Ok(pt)
    }
}

fn generate_key() -> [u8; 32] {
    use rand::RngCore;
    let mut k = [0u8; 32];
    rand::rngs::OsRng.fill_bytes(&mut k);
    k
}

#[cfg(test)]
mod tests {
    use super::*;

    fn fixed_crypto() -> LocalCrypto {
        // Deterministic test ключ — НЕ для production.
        LocalCrypto::with_key([42u8; 32])
    }

    #[test]
    fn round_trip_seals_and_opens() {
        let c = fixed_crypto();
        let pt = b"{\"hello\":\"world\"}";
        let sealed = c.seal(pt).unwrap();
        assert!(
            sealed.len() > NONCE_LEN + pt.len(),
            "sealed output must include nonce + ciphertext + auth tag"
        );
        let opened = c.open(&sealed).unwrap();
        assert_eq!(opened, pt);
    }

    #[test]
    fn each_seal_uses_unique_nonce() {
        let c = fixed_crypto();
        let a = c.seal(b"same").unwrap();
        let b = c.seal(b"same").unwrap();
        // С разными nonce одинаковый plaintext даёт разный ciphertext —
        // защита от частотного анализа повторяющихся событий.
        assert_ne!(a, b, "two seals of identical plaintext must differ");
    }

    #[test]
    fn tampered_ciphertext_fails_auth() {
        let c = fixed_crypto();
        let mut sealed = c.seal(b"important").unwrap();
        // Flip последнего байта (часть GCM tag) → decrypt должен отклонить.
        let last = sealed.len() - 1;
        sealed[last] ^= 0x01;
        assert!(c.open(&sealed).is_err());
    }

    #[test]
    fn wrong_key_fails_auth() {
        let a = LocalCrypto::with_key([1u8; 32]);
        let b = LocalCrypto::with_key([2u8; 32]);
        let sealed = a.seal(b"secret").unwrap();
        assert!(b.open(&sealed).is_err());
    }

    #[test]
    fn rejects_short_records() {
        let c = fixed_crypto();
        assert!(c.open(&[0u8; 5]).is_err());
        assert!(c.open(&[0u8; NONCE_LEN + 15]).is_err());
    }
}
