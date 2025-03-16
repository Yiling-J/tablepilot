import { TableCreateRequest } from "@/actions";
import { CreateTableDialog } from "@/components/dialog/create-table";
import { JSONObject } from "@/json";
import { ReactNode, createContext, useContext, useRef, useState } from "react";

interface CreateTableDialogContextValue {
  isOpen: boolean;
  openNewTableDialog: () => void;
  withForm: (form: TableCreateRequest) => void;
  clearForm: () => void;
  withRows: (rows: JSONObject[]) => void;
  clearRows: () => void;
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
  const [form, setForm] = useState<TableCreateRequest | undefined>(undefined);
  const withForm = (form: TableCreateRequest) => setForm(form);
  const clearForm = () => setForm(undefined);
  const rowsRef = useRef<JSONObject[] | undefined>(undefined);
  const withRows = (rows: JSONObject[]) => {
    rowsRef.current = rows;
  };
  const clearRows = () => {
    rowsRef.current = undefined;
  };

  return (
    <CreateTableDialogContext.Provider
      value={{
        isOpen,
        openNewTableDialog,
        withForm,
        clearForm,
        withRows,
        clearRows,
      }}
    >
      {children}
      <CreateTableDialog
        isOpen={isOpen}
        setIsOpen={setIsOpen}
        close={closeDialog}
        form={form}
        rows={rowsRef.current}
      />
    </CreateTableDialogContext.Provider>
  );
}
