import { CreateTableDialog } from "@/components/dialog/create-table";
import { ReactNode, createContext, useContext, useState } from "react";

interface CreateTableDialogContextValue {
  isOpen: boolean;
  openNewTableDialog: () => void;
}

const CreateTableDialogContext = createContext<
  CreateTableDialogContextValue | undefined
>(undefined);

export function useCreateTableDialog() {
  const context = useContext(CreateTableDialogContext);
  if (!context) {
    throw new Error(
      "useCreateTableDialog must be used within a CreateTableDialogProvider",
    );
  }
  return context;
}

interface CreateTableDialogProviderProps {
  children: ReactNode;
}

export function CreateTableDialogProvider({
  children,
}: CreateTableDialogProviderProps) {
  const [isOpen, setIsOpen] = useState(false);

  const openNewTableDialog = () => setIsOpen(true);
  const closeDialog = () => setIsOpen(false);

  return (
    <CreateTableDialogContext.Provider value={{ isOpen, openNewTableDialog }}>
      {children}
      <CreateTableDialog
        isOpen={isOpen}
        setIsOpen={setIsOpen}
        close={closeDialog}
      />
    </CreateTableDialogContext.Provider>
  );
}
