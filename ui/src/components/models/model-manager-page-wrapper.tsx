import { ModeToggle } from "../darkmode";
import { TablepilotHeader } from "../header";
import { ScrollArea } from "../ui/scroll-area";
import { ModelManager } from "./model-manager";

export function ModelManagerPageWrapper() {
  return (
    <div className="grow overflow-auto h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="models" />
      <ScrollArea className="h-[calc(100vh-120px)]">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 w-full tab-content-container">
          <div className="space-y-8">
            <ModelManager />
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}
