type NavbarProps = {
  title: string;
};

export function Navbar({ title }: NavbarProps) {
  return (
    <header className="rounded-xl bg-linear-to-r from-primary-strong to-primary px-4 py-3 text-header-text dark:from-primary-strong dark:to-primary dark:text-header-text">
      <h1 className="m-0 text-[clamp(1.25rem,2.2vw,1.75rem)] tracking-[0.03em]">{title}</h1>
    </header>
  );
}
