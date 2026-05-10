// @eop/ui сам — авторы Button/IconButton/SecretField, поэтому правила
// no-restricted-syntax на нас же бессмысленны. Базовый config держим
// для единообразия инфраструктуры (lint scripts, ts-eslint парсер),
// но restricted-syntax выключаем именно в этом пакете.
import { baseConfig } from "./eslint.base.js";

export default baseConfig({ extraRules: { "no-restricted-syntax": "off" } });
