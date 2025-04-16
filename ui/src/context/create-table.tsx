import { TableCreateRequest } from "@/actions";
import { CreateTableDialog } from "@/components/dialog/create-table";
import { JSONObject } from "@/json";
import { ReactNode, createContext, useContext, useRef, useState } from "react";

interface CreateTableDialogContextValue {
  isOpen: boolean;
  openNewTableDialog: () => void;
  withForm: (form: TableCreateRequest) => void;
  withRows: (rows: JSONObject[]) => void;
  withTable: (table: string) => void;
  withSubmitCallback: (callback: () => Promise<void>) => void;
  clear: () => void;
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

interface submitCallback {
  fn?: () => Promise<void>;
}

export function CreateTableDialogProvider({
  children,
}: CreateTableDialogProviderProps) {
  const [isOpen, setIsOpen] = useState(false);
  const openNewTableDialog = () => setIsOpen(true);
  const closeDialog = () => setIsOpen(false);
  const [form, setForm] = useState<TableCreateRequest | undefined>(undefined);
  const [table, setTable] = useState<string | undefined>(undefined);
  const submitCallbackRef = useRef<submitCallback>({});
  const withForm = (form: TableCreateRequest) => setForm(form);
  const withTable = (table: string) => setTable(table);
  const rowsRef = useRef<JSONObject[] | undefined>(undefined);
  const withRows = (rows: JSONObject[]) => {
    rowsRef.current = rows;
  };
  const withSubmitCallback = (callback: () => Promise<void>) => {
    submitCallbackRef.current.fn = callback;
  };

  const clear = () => {
    setForm(undefined);
    setTable(undefined);
    rowsRef.current = undefined;
    submitCallbackRef.current.fn = undefined;
  };

  return (
    <CreateTableDialogContext.Provider
      value={{
        isOpen,
        openNewTableDialog,
        withForm,
        withRows,
        withTable,
        withSubmitCallback,
        clear,
      }}
    >
      {children}
      <CreateTableDialog
        table={table}
        isOpen={isOpen}
        setIsOpen={(v: boolean) => {
          if (!v) {
            clear();
          }
          setIsOpen(v);
        }}
        close={closeDialog}
        form={form}
        rows={rowsRef.current}
        submitCallback={submitCallbackRef.current.fn}
      />
    </CreateTableDialogContext.Provider>
  );
}
