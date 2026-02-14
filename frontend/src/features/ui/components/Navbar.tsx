type NavbarProps = {
  title: string;
};

export function Navbar({ title }: NavbarProps) {
  return (
    <header className="app-header">
      <h1>{title}</h1>
    </header>
  );
}
