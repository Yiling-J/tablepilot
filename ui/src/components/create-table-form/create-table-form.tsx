import {
    TableCreateRequest,
    createRows,
    createTable,
    updateTable,
} from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useCreateTableDialog } from "@/context/create-table";
import { useTables } from "@/context/tables";
import { JSONObject } from "@/json";
import { cn } from "@/lib/utils";
import { ReloadIcon } from "@radix-ui/react-icons";
import { useState } from "react";
import { toast } from "react-hot-toast";
import { useNavigate } from "react-router-dom";
import { ColumnsForm } from "./columns-form";
import { JsonPreview } from "./json-preview";
import { NameDescriptionForm } from "./name-description-form";
import { SourcesForm } from "./sources-form";

const initialFormData: TableCreateRequest = {
  name: "",
  description: "",
  sources: [],
  columns: [],
};

interface CreateTableFormProps {
  table?: string;
  close: () => void;
  form?: TableCreateRequest;
  rows?: JSONObject[];
  submitCallback?: () => Promise<void>;
}

export default function CreateTableForm({
  table,
  close,
  form,
  rows,
  submitCallback,
}: CreateTableFormProps) {
  const [formData, setFormData] = useState<TableCreateRequest>(
    form ?? initialFormData,
  );
  const [loading, setLoading] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState("step1");
  const [showPreview, setShowPreview] = useState(false);
  const { clear } = useCreateTableDialog();
  const navigate = useNavigate();
  const { refreshTables } = useTables();

  const updateFormData = (data: Partial<TableCreateRequest>) => {
    setFormData((prev) => ({ ...prev, ...data }));
  };

  const handleNext = () => {
    if (activeTab === "step1") setActiveTab("step2");
    else if (activeTab === "step2") setActiveTab("step3");
  };

  const handlePrevious = () => {
    if (activeTab === "step3") setActiveTab("step2");
    else if (activeTab === "step2") setActiveTab("step1");
  };

  const handleSubmit = async () => {
    setLoading(true);
    let info;
    try {
      if (table !== undefined) {
        info = await updateTable(table, formData);
      } else {
        info = await createTable(formData);
      }
      if (rows && rows.length > 0) {
        await createRows(info.id, rows);
      }
      if (submitCallback) {
        await submitCallback();
      }
      await refreshTables();
      close();
      clear();
      navigate(`/tables/${info.id}`);
    } catch (err) {
      console.log("create table failed: ", err);
      toast.error("Creation failed. Please wait and try again.");
    } finally {
      setLoading(false);
    }
  };

  const isStep1Valid = formData.name.trim() !== "";
  const isStep2Valid = true;
  const isStep3Valid = formData.columns.length > 0;

  return (
    <div className="mx-0">
      <div
        className={cn(
          "container mx-auto py-2 px-2 overflow-y-auto scrollbar-thin grow pl-3",
          showPreview ? "max-h-[35vh]" : "max-h-[65vh]",
        )}
      >
        <Card className="w-full max-w-4xl mx-auto">
          <CardHeader>
            <CardTitle>Table Configuration</CardTitle>
            <CardDescription>
              {table === undefined
                ? "Create your table configuration or import JSON"
                : "Update your table configuration or import JSON"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs
              value={activeTab}
              onValueChange={setActiveTab}
              className="w-full"
            >
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="step1">Basic</TabsTrigger>
                <TabsTrigger value="step2" disabled={!isStep1Valid || loading}>
                  Sources
                </TabsTrigger>
                <TabsTrigger
                  value="step3"
                  disabled={!isStep1Valid || !isStep2Valid || loading}
                >
                  Columns
                </TabsTrigger>
              </TabsList>
              <TabsContent value="step1">
                <NameDescriptionForm
                  formData={formData}
                  updateFormData={updateFormData}
                />
              </TabsContent>
              <TabsContent value="step2">
                <SourcesForm
                  formData={formData}
                  updateFormData={updateFormData}
                />
              </TabsContent>
              <TabsContent value="step3">
                <ColumnsForm
                  formData={formData}
                  updateFormData={updateFormData}
                  disabled={!isStep1Valid || loading}
                />
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>

      <div className="flex w-full justify-between my-2">
        <Button
          variant="outline"
          onClick={handlePrevious}
          disabled={activeTab === "step1"}
        >
          Previous
        </Button>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => setShowPreview(!showPreview)}
          >
            {showPreview ? "Hide" : "Show"} JSON
          </Button>
          {activeTab !== "step3" ? (
            <Button
              onClick={handleNext}
              disabled={
                (activeTab === "step1" && !isStep1Valid) ||
                (activeTab === "step2" && !isStep2Valid)
              }
            >
              Next
            </Button>
          ) : (
            <Button onClick={handleSubmit} disabled={loading || !isStep3Valid}>
              {loading ? <ReloadIcon className="animate-spin" /> : "Complete"}
            </Button>
          )}
        </div>
      </div>

      {showPreview && (
        <div className="max-h-[35vh] scrollbar-thin grow overflow-auto pl-3">
          <Card className="w-full max-w-4xl mx-auto mt-6">
            <CardHeader>
              <CardTitle>JSON Preview</CardTitle>
            </CardHeader>
            <CardContent>
              <JsonPreview formData={formData} />
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
