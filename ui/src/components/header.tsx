import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useSidebar } from "@/context/sidebar";
import { cn } from "@/lib/utils";
import { ChevronRightIcon, ReloadIcon } from "@radix-ui/react-icons";
import { useState } from "react";
import { IconGithub } from "./ui/icons";

function ModeSwitch({
  modeRef,
  disabled,
}: {
  modeRef: React.MutableRefObject<"generate" | "autofill">;
  disabled: boolean;
}) {
  const [mode, setMode] = useState<"generate" | "autofill">("generate");

  return (
    <div className="flex items-center justify-center pr-12 rounded-xl">
      <div className="relative flex py-1 rounded-full bg-gradient-to-r from-yellow-500/10 to-amber-500/10 border border-purple-500/30 dark:from-cyan-500/20 dark:to-purple-500/20">
        <div
          className={cn(
            "absolute h-full top-0 w-1/2 rounded-full transition-all duration-200 ease-out z-0",
            "bg-gradient-to-r from-yellow-500/80 to-amber-500/80 dark:from-cyan-600/80 dark:to-purple-600/80",
            mode === "autofill" ? "translate-x-full" : "translate-x-0",
          )}
        />

        <button
          disabled={disabled}
          onClick={() => {
            modeRef.current = "generate";
            setMode("generate");
          }}
          className={cn(
            "relative z-10 pr-5 pl-3 py-0 text-xs transition-colors duration-100",
            "focus:outline-none focus:ring-purple-500/50 focus:ring-offset-2 focus:ring-offset-gray-900",
            mode === "generate"
              ? "text-forground"
              : "dark:text-gray-400 text-gray-500 hover:text-gray-200",
          )}
        >
          Generate
        </button>

        <button
          disabled={disabled}
          onClick={() => {
            modeRef.current = "autofill";
            setMode("autofill");
          }}
          className={cn(
            "relative z-10 pr-5 pl-3 py-0 text-xs transition-colors duration-100",
            "focus:outline-none focus:ring-purple-500/50 focus:ring-offset-2 focus:ring-offset-gray-900",
            mode === "autofill"
              ? "text-forground"
              : "dark:text-gray-400 text-gray-500 hover:text-gray-200",
          )}
        >
          Autofill
        </button>
      </div>
    </div>
  );
}

interface TablepilotHeaderProps {
  title: string;
  modeRef?: React.MutableRefObject<"generate" | "autofill">;
  modeSwitchDisabled?: boolean;
  currentTab?: string;
  onTabChange?: (tabName: string) => void;
  onRefresh?: () => void;
}

export function TablepilotHeader({
  title,
  modeRef,
  modeSwitchDisabled,
  currentTab,
  onTabChange,
  onRefresh,
}: TablepilotHeaderProps) {
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
      <header className="sticky top-0 font-bold flex items-center gap-4 border-b bg-background px-4 md:px-6 justify-between py-2">
        <div className="flex items-center text-xl tracking-wider">
          {collapsed && (
            <ChevronRightIcon
              className="mr-2 cursor-pointer"
              onClick={() => setCollapsed(false)}
            />
          )}
          {title}
        </div>
        <div className="relative flex items-center">
          {currentTab && onTabChange && (
            <div className="flex items-center">
              <Button
                variant="ghost"
                className="rounded-full"
                onClick={() => onTabChange("tables")}
              >
                <h1
                  className={cn(
                    "text-xl font-bold tracking-wider",
                    currentTab !== "tables" && "text-primary/25",
                  )}
                >
                  Tables
                </h1>
              </Button>
              <p className="mx-2 text-xl font-bold">/</p>
              <Button
                variant="ghost"
                className="rounded-full"
                onClick={() => onTabChange("workflows")}
              >
                <h1
                  className={cn(
                    "text-xl font-bold tracking-wider",
                    currentTab !== "workflows" && "text-primary/25",
                  )}
                >
                  Workflows
                </h1>
              </Button>
            </div>
          )}
          {onRefresh && (
            <Button
              variant="outline"
              onClick={onRefresh}
              className="ml-4 mr-4" // Added margin for spacing
            >
              <ReloadIcon className="w-6 h-6 mr-2" />
              Refresh
            </Button>
          )}
          {modeRef && (
            <ModeSwitch
              modeRef={modeRef}
              disabled={modeSwitchDisabled ?? true}
            />
          )}
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
