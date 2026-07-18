interface ProjectHeaderProps {
  children: React.ReactNode;
}

export const ProjectHeader = ({ children }: ProjectHeaderProps) => {
  return <header className="border-border bg-card flex min-h-16 w-full items-center border-b px-4">{children}</header>;
};

const HeaderLogo = () => {
  return <img src="/logo.png" className="h-8 w-8 object-contain" />;
};

ProjectHeader.Logo = HeaderLogo;
