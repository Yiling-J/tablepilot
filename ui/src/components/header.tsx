import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useSidebar } from "@/context/sidebar";
import { cn } from "@/lib/utils";
import { ChevronRightIcon } from "@radix-ui/react-icons";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { IconGithub } from "./ui/icons";

interface TablepilotHeaderProps {
  title: string;
  modeRef?: React.MutableRefObject<"generate" | "autofill">;
  modeSwitchDisabled?: boolean;
  currentTab?: string;
  onTabChange?: (tabName: string) => void;
}

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

export function TablepilotHeader({
  title,
  currentTab,
  modeRef,
  modeSwitchDisabled,
}: TablepilotHeaderProps) {
  const { collapsed, setCollapsed } = useSidebar();
  const navigate = useNavigate();

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
        {/* Left Group (Title & Sidebar Toggle, THEN Tabs) */}
        <div className="flex items-center gap-x-6">
          {/* Title and Sidebar Toggle */}
          <div className="flex items-center text-xl tracking-wider font-bold">
            {collapsed && (
              <ChevronRightIcon
                className="mr-2 cursor-pointer"
                onClick={() => setCollapsed(false)}
              />
            )}
            {title}
          </div>

          {/* Tables/Workflows switch (only if currentTab and onTabChange are provided) */}
          {currentTab && (
            <div className="flex items-center">
              {" "}
              {/* Container for the two tabs */}
              <div
                onClick={() => navigate("/tables")}
                className="cursor-pointer"
              >
                <h1
                  className={cn(
                    "text-base font-medium py-2 px-3 border-b-2",
                    currentTab === "tables"
                      ? "border-primary text-primary"
                      : "text-muted-foreground hover:text-primary hover:border-primary border-transparent",
                  )}
                >
                  Tables
                </h1>
              </div>
              {/* Datasets Tab */}
              <div
                onClick={() => navigate("/datasets")}
                className="ml-4 cursor-pointer"
              >
                <h1
                  className={cn(
                    "text-base font-medium py-2 px-3 border-b-2",
                    currentTab === "datasets"
                      ? "border-primary text-primary"
                      : "text-muted-foreground hover:text-primary hover:border-primary border-transparent",
                  )}
                >
                  Datasets
                </h1>
              </div>
              {/* Workflows Tab */}
              <div
                onClick={() => navigate("/workflows")}
                className="ml-4 cursor-pointer"
              >
                <h1
                  className={cn(
                    "text-base font-medium py-2 px-3 border-b-2",
                    currentTab === "workflows"
                      ? "border-primary text-primary"
                      : "text-muted-foreground hover:text-primary hover:border-primary border-transparent",
                  )}
                >
                  Workflows
                </h1>
              </div>
              {/* Models Tab */}
              <div
                onClick={() => navigate("/models")}
                className="ml-4 cursor-pointer"
              >
                <h1
                  className={cn(
                    "text-base font-medium py-2 px-3 border-b-2",
                    currentTab === "models"
                      ? "border-primary text-primary"
                      : "text-muted-foreground hover:text-primary hover:border-primary border-transparent",
                  )}
                >
                  Models
                </h1>
              </div>
            </div>
          )}
        </div>

        {/* Right Group (Controls: ModeSwitch, Avatar) */}
        <div className="relative flex items-center">
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
