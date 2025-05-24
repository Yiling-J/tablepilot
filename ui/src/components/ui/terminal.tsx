import { type ReactNode, type Ref } from "react";

type TerminalProps = {
  ref: Ref<HTMLDivElement>;
  children: ReactNode;
  running: boolean;
};

export function Terminal({ children, running, ref }: TerminalProps) {
  return (
    <div className="h-full flex flex-col">
      <div className="bg-gray-800 p-1 border-b border-gray-700 flex items-center gap-2">
        <div className="text-xs text-gray-400 ml-3 mt-1">Workflow Output</div>
      </div>
      <div
        ref={ref}
        className="flex-1 p-4 font-mono text-sm overflow-auto hide-scrollbar"
      >
        {children}
        {running && (
          <span className="inline-block text-white animate-[blink_1s_step-end_infinite]">
            _
          </span>
        )}
      </div>
    </div>
  );
}

Terminal.displayName = "Terminal";
