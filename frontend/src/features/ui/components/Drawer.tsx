import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

type DrawerProps = {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
  width?: number | string;
  positionClassName?: string;
};

function resolveDrawerWidth(width: DrawerProps['width']) {
  if (typeof width === 'number') {
    return `${width}px`;
  }

  return width ?? '32rem';
}

export function Drawer({ isOpen, onClose, children, width, positionClassName }: DrawerProps) {
  const portalRoot = typeof document === 'undefined' ? null : document.body;

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!portalRoot) {
    return null;
  }

  return createPortal(
    <div
      aria-hidden={!isOpen}
      className={`${positionClassName ?? 'absolute inset-y-4 right-full z-20 flex pr-4'} ${
        isOpen ? 'pointer-events-auto' : 'pointer-events-none'
      }`}
    >
      <section
        role="dialog"
        aria-modal="false"
        className={`shell-panel flex h-full min-w-lg flex-col rounded-xl border shadow-[0_24px_80px_-28px_rgba(0,0,0,0.85)] transition-all duration-300 ease-out ${
          isOpen
            ? 'pointer-events-auto translate-x-0 opacity-100'
            : 'pointer-events-none -translate-x-6 opacity-0'
        }`}
        style={{ width: resolveDrawerWidth(width) }}
      >
        <div className="flex items-center justify-end px-3 pt-3">
          <button
            type="button"
            onClick={onClose}
            className="shell-button cursor-pointer rounded-sm border px-2 py-1 text-xs font-medium transition-colors"
          >
            Close
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto px-4 pb-4">{children}</div>
      </section>
    </div>,
    portalRoot,
  );
}
