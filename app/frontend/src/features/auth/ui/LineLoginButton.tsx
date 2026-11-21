import type { CSSProperties } from "react";
import lineIcon from "../line-icon/line_88.png";

type LineLoginButtonProps = {
  onLogin: () => void;
  loggingIn: boolean;
};

// The LINE icon from the LINE Login button template (DeskTop/2x/44dp), shown at
// 44 px. It is used as a mask so it takes the color the guidelines give each
// state, and is never redrawn or stretched.
const iconMask: CSSProperties = {
  maskImage: `url(${lineIcon})`,
  maskSize: "44px 44px",
  maskRepeat: "no-repeat",
};

// The LINE Login button (A1, A5). Its colors, icon, separator, and wording
// follow the LINE Login button design guidelines
// (https://developers.line.biz/ja/docs/line-login/login-button/), not
// tasuki's buttons (docs/design/components.md):
// - base #06C755; hover and pressed put black at 10 % and 30 % on it, under
//   the icon, the separator, and the text;
// - icon and text #FFFFFF; separator black at 8 %;
// - disabled: white with a #E5E5E5 border at 60 %, icon and text #1E1E1E at
//   20 %, separator #E5E5E5 at 60 %.
export function LineLoginButton({ onLogin, loggingIn }: LineLoginButtonProps) {
  return (
    <button
      type="button"
      onClick={onLogin}
      disabled={loggingIn}
      className="group relative flex min-h-11 w-full items-stretch overflow-hidden rounded-lg bg-[#06C755] text-white before:absolute before:inset-0 before:bg-black/0 hover:before:bg-black/10 active:before:bg-black/30 disabled:border disabled:border-[#E5E5E5]/60 disabled:bg-white disabled:text-[#1E1E1E]/20 disabled:before:bg-transparent"
    >
      <span
        aria-hidden="true"
        style={iconMask}
        className="relative size-11 shrink-0 bg-white group-disabled:bg-[#1E1E1E]/20"
      />
      <span
        aria-hidden="true"
        className="relative w-px shrink-0 bg-black/8 group-disabled:bg-[#E5E5E5]/60"
      />
      <span className="relative flex grow items-center justify-center px-8 text-headline font-bold whitespace-nowrap">
        LINEでログイン
      </span>
    </button>
  );
}
