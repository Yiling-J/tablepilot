import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useSidebar } from "@/context/sidebar";
import { ChevronRightIcon } from "@radix-ui/react-icons";
import { IconGithub } from "./ui/icons";

interface TablepilotHeaderProps {
  title: string;
}

export function TablepilotHeader({ title }: TablepilotHeaderProps) {
  const { collapsed, setCollapsed } = useSidebar();

  const handleAvatarClick = () => {
    window.open(
      "https://github.com/Yiling-J/tablepilot",
      "_blank",
      "noopener,noreferrer",
    );
  };

  return (
    <div>
      <header className="sticky top-0 flex items-center gap-4 border-b bg-background px-4 md:px-6 justify-between py-2">
        <div className="flex items-center text-xl tracking-wider">
          {collapsed && (
            <ChevronRightIcon
              className="mr-2 cursor-pointer"
              onClick={() => setCollapsed(false)}
            />
          )}
          {title}
        </div>
        <div className="relative">
          <Avatar
            onClick={() => {
              handleAvatarClick();
            }}
            className="cursor-pointer"
          >
            <AvatarFallback>
              <IconGithub className="size-fit" />
            </AvatarFallback>
          </Avatar>
        </div>
      </header>
    </div>
  );
}
