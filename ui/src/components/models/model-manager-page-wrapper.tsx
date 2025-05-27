import { useState } from 'react';
import { ModelManager } from "./model-manager";
import { TablepilotHeader } from "../header";
import { PlusCircledIcon } from "@radix-ui/react-icons";
// import { ModeToggle } from "../darkmode"; // Not including for now, can be added if desired

export function ModelManagerPageWrapper() {
  const [isAddProviderDialogOpen, setIsAddProviderDialogOpen] = useState(false);

  const openAddProviderDialog = () => {
    setIsAddProviderDialogOpen(true);
  };

  const handleAddProviderDialogDismiss = () => {
    setIsAddProviderDialogOpen(false);
  };

  return (
    <div className="grow overflow-auto h-full flex flex-col">
      {/* <ModeToggle hide={true} /> */}
      <TablepilotHeader title="Tablepilot" currentTab="models" />
      <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 w-full"> {/* Added w-full for consistency */}
        {/* Using a similar container structure as TableListPage */}
        <div className="space-y-8"> {/* Assuming space-y-8 is desired, like in TableListPage */}
          <ModelManager
            searchTerm=""
            shouldOpenAddProviderDialog={isAddProviderDialogOpen}
            onAddProviderDialogDismiss={handleAddProviderDialogDismiss}
          />
        </div>
      </div>
      <button
        onClick={openAddProviderDialog}
        className="fixed bottom-8 right-8 bg-green-500 hover:bg-green-600 text-white p-4 rounded-full shadow-lg"
        aria-label="Add provider"
      >
        <PlusCircledIcon className="h-6 w-6" />
      </button>
    </div>
  );
}
