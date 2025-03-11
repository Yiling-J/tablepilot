import { useState, useEffect } from "react";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { cn } from "@/lib/utils";

// Define theme options
type Theme = "light" | "dark" | "system";

interface ModeToggleProps {
  hide: boolean;
}

export function ModeToggle({ hide }: ModeToggleProps) {
  // Get initial theme from localStorage or use "system" as default
  const [theme, setTheme] = useState<Theme>(() => {
    const storedTheme = localStorage.getItem("theme") as Theme;
    return storedTheme || "dark";
  });

  useEffect(() => {
    const root = document.documentElement;

    // Function to apply theme
    const applyTheme = (theme: Theme) => {
      if (theme === "dark") {
        root.classList.add("dark");
      } else {
        root.classList.remove("dark");
      }
    };

    // Apply theme on initial render and when theme changes
    applyTheme(theme);

    // Update localStorage whenever the theme changes
    localStorage.setItem("theme", theme);
  }, [theme]);

  const handleThemeChange = (value: string) => {
    if (["light", "dark", "system"].includes(value)) {
      setTheme(value as Theme);
    } else {
      // Handle invalid value (optional)
      console.error("Invalid theme value:", value);
    }
  };

  if (hide) {
    return <div hidden={true}></div>;
  }

  return (
    <div className="p-1 cursor-pointer max-w-[160px]">
      <ToggleGroup
        aria-label="Toggle theme"
        value={theme}
        onValueChange={handleThemeChange}
        className="flex items-center justify-start pl-5 bg-transparent"
        type="single"
      >
        <ToggleGroupItem
          value="dark"
          className="px-1.5 bg-transparent data-[state=on]:bg-transparent transition-none"
        >
          <MoonIcon
            className={cn("h-4 w-4", theme != "dark" ? "text-gray-400" : "")}
          />
        </ToggleGroupItem>
        <ToggleGroupItem
          value="light"
          className="px-1.5 bg-transparent data-[state=on]:bg-transparent transition-none"
        >
          <SunIcon
            className={cn("h-4 w-4", theme != "light" ? "text-gray-400" : "")}
          />
        </ToggleGroupItem>
        <ToggleGroupItem
          value="system"
          className="px-1.5 bg-transparent data-[state=on]:bg-transparent transition-none"
        >
          <ComputerIcon
            className={cn("h-4 w-4", theme != "system" ? "text-gray-400" : "")}
          />
        </ToggleGroupItem>
      </ToggleGroup>
    </div>
  );
}

function ComputerIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect width="14" height="8" x="5" y="2" rx="2" />
      <rect width="20" height="8" x="2" y="14" rx="2" />
      <path d="M6 18h2" />
      <path d="M12 18h6" />
    </svg>
  );
}

function MoonIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
    </svg>
  );
}

function SunIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2" />
      <path d="M12 20v2" />
      <path d="m4.93 4.93 1.41 1.41" />
      <path d="m17.66 17.66 1.41 1.41" />
      <path d="M2 12h2" />
      <path d="M20 12h2" />
      <path d="m6.34 17.66-1.41 1.41" />
      <path d="m19.07 4.93-1.41 1.41" />
    </svg>
  );
}
