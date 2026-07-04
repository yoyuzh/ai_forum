import { useEffect } from "react";
import MaterialIcon from "./MaterialIcon";

type AlertTone = "success" | "error" | "info";

interface AlertBarProps {
  tone: AlertTone;
  message: string;
  onClose?: () => void;
  /** Auto-dismiss after ms. 0 = persistent. */
  duration?: number;
}

const TONE_STYLES: Record<AlertTone, { wrap: string; icon: string; iconName: string }> = {
  success: {
    wrap: "bg-cohere-success text-cohere-deep-green border-cohere-secondary/40",
    icon: "text-cohere-secondary",
    iconName: "check_circle",
  },
  error: {
    wrap: "bg-cohere-error-container text-cohere-error border-cohere-error/30",
    icon: "text-cohere-error",
    iconName: "error",
  },
  info: {
    wrap: "bg-cohere-pale-blue text-cohere-info-blue border-cohere-info-blue/30",
    icon: "text-cohere-info-blue",
    iconName: "info",
  },
};

/**
 * Inline alert strip for form feedback.
 * Self-dismissing via `duration`. No global store — each page owns its own
 * instance, keeping the change surface small per web/AGENTS.md.
 */
export default function AlertBar({ tone, message, onClose, duration = 0 }: AlertBarProps) {
  const styles = TONE_STYLES[tone];

  useEffect(() => {
    if (duration > 0 && onClose) {
      const t = setTimeout(onClose, duration);
      return () => clearTimeout(t);
    }
  }, [duration, onClose]);

  return (
    <div
      role="status"
      aria-live="polite"
      className={`flex items-start gap-sm rounded-sm border px-md py-sm font-caption ${styles.wrap}`}
    >
      <MaterialIcon name={styles.iconName} className={`mt-xxs ${styles.icon}`} size={18} />
      <span className="flex-1 leading-relaxed">{message}</span>
      {onClose && (
        <button
          type="button"
          onClick={onClose}
          aria-label="关闭提示"
          className={`shrink-0 rounded-xs px-xs transition-colors hover:bg-black/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue ${styles.icon}`}
        >
          <MaterialIcon name="close" size={16} />
        </button>
      )}
    </div>
  );
}
