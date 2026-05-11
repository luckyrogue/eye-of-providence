// 3D eye stage с орбитальными нодами, pyramid, dilating pupil, scanline.
// Чистый CSS-animation — никакого WebGL/canvas, всё в transform 3D + keyframes
// (см. [data-theme="eop"] keyframes в ui/src/shared/styles.css).
//
// Подложка mouse-tracking: parent <div> rotateX/Y по перемещению мыши даёт
// иллюзию "следит за курсором" — детали в useEffect.

import { useEffect, useRef, type ReactElement } from "react";

type IconName = "VSCode" | "Browser" | "Terminal" | "Anthropic" | "Cursor" | "Copilot";

const Icons: Record<IconName, ReactElement> = {
  VSCode: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M17 3 7 11 3 8v8l4-3 10 8 4-2V5l-4-2Z" />
      <path d="M17 3 7 13" />
    </svg>
  ),
  Browser: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18" />
    </svg>
  ),
  Terminal: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="m6 9 3 3-3 3M12 16h6" />
    </svg>
  ),
  Anthropic: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M14.5 4h3.7L24 20h-3.7l-1.3-3.4h-6.8L10.9 20H7.2L13 4h1.5zm-.7 3.7-2.4 6.4h4.8l-2.4-6.4zM3.7 4h3.7L13 20H9.3l-5.6-16z" />
    </svg>
  ),
  Cursor: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M4 4l16 8-7 1-1 7L4 4Z" />
    </svg>
  ),
  Copilot: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
      <path d="M4 12a8 8 0 0 1 16 0v3a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4v-3Z" />
      <circle cx="9" cy="13" r="1.5" fill="currentColor" />
      <circle cx="15" cy="13" r="1.5" fill="currentColor" />
    </svg>
  ),
};

const NODES: { icon: IconName; angle: number; glow?: boolean }[] = [
  { icon: "VSCode", angle: 0, glow: true },
  { icon: "Browser", angle: 60 },
  { icon: "Terminal", angle: 120, glow: true },
  { icon: "Anthropic", angle: 180 },
  { icon: "Cursor", angle: 240, glow: true },
  { icon: "Copilot", angle: 300 },
];

export function EyeScene() {
  const stageRef = useRef<HTMLDivElement>(null);
  // Mouse-tracking: tilt stage по cursor position для "следящего" эффекта.
  // Throttling не нужен — CSS transition в 0.3s сам сглаживает.
  useEffect(() => {
    const el = stageRef.current;
    if (!el) return;
    const onMove = (e: MouseEvent) => {
      const r = el.getBoundingClientRect();
      const x = (e.clientX - r.left - r.width / 2) / r.width;
      const y = (e.clientY - r.top - r.height / 2) / r.height;
      el.style.setProperty("--mx", `${-y * 12}deg`);
      el.style.setProperty("--my", `${x * 12}deg`);
    };
    window.addEventListener("mousemove", onMove);
    return () => window.removeEventListener("mousemove", onMove);
  }, []);

  return (
    <div
      className="eye-stage"
      ref={stageRef}
      style={{
        transform: "rotateX(var(--mx,0)) rotateY(var(--my,0))",
        transition: "transform 0.3s ease-out",
      }}
    >
      <div className="eye-orbit">
        <div className="ring r1">
          <span className="ring-tick" />
        </div>
        <div className="ring r2" />
        <div className="ring r3">
          <span className="ring-tick" />
        </div>
        {NODES.map((n, i) => {
          const radius = 47;
          const rad = (n.angle * Math.PI) / 180;
          const x = Math.cos(rad) * radius;
          const y = Math.sin(rad) * radius;
          return (
            <div
              key={i}
              className={`orbit-node ${n.glow ? "glow" : ""}`}
              style={{
                transform: `translate(-50%, -50%) translate(${x}%, ${y}%)`,
              }}
            >
              {Icons[n.icon]}
            </div>
          );
        })}
      </div>
      <div className="eye-core">
        <div className="pyramid">
          <div className="pyramid-face" />
          <div className="pyramid-face f2" />
          <div className="pyramid-face f3" />
        </div>
        <div className="eye-pupil" />
      </div>
      <div className="scanline" />
    </div>
  );
}
