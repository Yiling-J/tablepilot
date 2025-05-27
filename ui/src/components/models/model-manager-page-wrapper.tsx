import { ModelManager } from "./model-manager";
import { TablepilotHeader } from "../header";
// import { ModeToggle } from "../darkmode"; // Not including for now, can be added if desired

export function ModelManagerPageWrapper() {
  return (
    <div className="grow overflow-auto h-full flex flex-col">
      {/* <ModeToggle hide={true} /> */}
      <TablepilotHeader title="Tablepilot" currentTab="models" />
      <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 w-full"> {/* Added w-full for consistency */}
        {/* Using a similar container structure as TableListPage */}
        <div className="space-y-8"> {/* Assuming space-y-8 is desired, like in TableListPage */}
          <ModelManager searchTerm="" />
        </div>
      </div>
    </div>
  );
}
