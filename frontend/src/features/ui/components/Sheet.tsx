import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

type SheetProps = {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
  width?: number | string;
};

type SheetSectionProps = {
  children: ReactNode;
  className?: string;
};

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

function resolveSheetWidth(width: SheetProps['width']) {
  if (typeof width === 'number') {
    return `${width}px`;
  }

  return width ?? '42rem';
}

export function Sheet({ isOpen, onClose, children, width }: SheetProps) {
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
      className={`fixed inset-0 z-30 flex items-end justify-center p-4 transition-opacity duration-300 sm:p-6 ${
        isOpen ? 'pointer-events-auto opacity-100' : 'pointer-events-none opacity-0'
      }`}
    >
      <button
        type="button"
        aria-label="Close sheet"
        onClick={onClose}
        className="absolute inset-0 border-0 bg-black/45 p-0"
        tabIndex={isOpen ? 0 : -1}
      />

      <section
        role="dialog"
        aria-modal="true"
        className={`shell-panel relative z-10 flex max-h-[min(54rem,calc(100dvh-1rem))] w-full flex-col overflow-hidden rounded-xl border shadow-[0_-24px_80px_-28px_rgba(0,0,0,0.85)] transition-transform duration-300 ease-out ${
          isOpen ? 'translate-y-0' : 'translate-y-[calc(100%+2rem)]'
        }`}
        style={{ maxWidth: resolveSheetWidth(width) }}
      >
        {children}
      </section>
    </div>,
    portalRoot,
  );
}

export function SheetHeader({ children, className }: SheetSectionProps) {
  return <div className={mergeClassNames('px-4 pb-3 pt-4 sm:px-5', className)}>{children}</div>;
}

export function SheetBody({ children, className }: SheetSectionProps) {
  return (
    <div className={mergeClassNames('min-h-0 flex-1 overflow-y-auto px-4 pb-6 pt-4 sm:px-5', className)}>
      {children}
    </div>
  );
}

export function SheetFooter({ children, className }: SheetSectionProps) {
  return (
    <div
      className={mergeClassNames(
        'shell-panel sticky bottom-0 z-10 border-t border-[color:var(--color-shell-border)] px-4 py-4 sm:px-5',
        className,
      )}
    >
      {children}
    </div>
  );
}
