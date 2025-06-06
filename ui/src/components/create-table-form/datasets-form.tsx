import {
    createDataset,
    CreateDatasetRequest,
    DatasetInfo,
    getTableDatasets,
    TableCreateRequest,
} from "@/actions";
import { CreateDatasetDialog } from "@/components/dialog/dataset/dataset";
import { DatasetPreviewDialog } from "@/components/dialog/dataset/preview";
import { Button } from "@/components/ui/button";
import { ContextVariable } from "@/components/ui/var-input";
import { useEffect, useState } from "react";
import { toast } from "react-hot-toast";

interface DatasetsFormProps {
  form?: TableCreateRequest;
  table?: string;
  variables?: ContextVariable[];
}

export function DatasetsForm({ table }: DatasetsFormProps) {
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [selectedDataset, setSelectedDataset] = useState<DatasetInfo | null>(
    null,
  );
  const [showDatasetDialog, setShowDatasetDialog] = useState<boolean>(false);

  const fetchDatasets = async () => {
    if (!table) return;
    setLoading(true);
    try {
      const response = await getTableDatasets(table);
      setDatasets(response.datasets);
    } catch (error) {
      console.error("Failed to fetch datasets:", error);
      toast.error("Failed to load datasets.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (table) {
      fetchDatasets();
    }
  }, [table]);

  const handleDatasetClick = (dataset: DatasetInfo) => {
    setSelectedDataset(dataset);
  };

  const handleAddDataset = (): void => {
    setShowDatasetDialog(true);
  };

  const handleCreateDataset = async (datasetData: {
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[];
    files?: File[];
  }) => {
    setLoading(true);
    try {
      const requestData: CreateDatasetRequest = {
        name: datasetData.name,
        description: datasetData.description,
        type: datasetData.type,
        data: datasetData.options || [],
        files: datasetData.files || [],
        private: true,
        table: table,
      };
      const id = await createDataset(requestData);
      toast.success("Dataset created successfully!");
      if (table) {
        fetchDatasets();
      } else {
        setDatasets((old) => [
          {
            id: id,
            name: datasetData.name,
            description: datasetData.description,
            type: datasetData.type,
            data: [],
          },
          ...old,
        ]);
      }
    } catch (error) {
      console.error("Failed to create dataset:", error);
      toast.error("Failed to create dataset.");
    } finally {
      setLoading(false);
      setShowDatasetDialog(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium">Datasets</h3>
        <Button onClick={handleAddDataset} variant="outline">
          Add Dataset
        </Button>
      </div>
      {loading && <p>Loading datasets...</p>}
      {!loading && datasets.length === 0 && (
        <p>No datasets found for this table.</p>
      )}
      {!loading && datasets.length > 0 && (
        <div>
          {datasets.map((dataset) => (
            <div
              key={dataset.id}
              onClick={() => handleDatasetClick(dataset)}
              className="cursor-pointer flex flex-row border px-2 py-4 justify-between mb-5 bg-secondary/50 rounded-md"
            >
              <div className="ml-4">{dataset.name}</div>
              <div className="mr-4">{dataset.type}</div>
            </div>
          ))}
        </div>
      )}

      {selectedDataset && (
        <DatasetPreviewDialog
          datasetId={selectedDataset.id}
          isOpen={!!selectedDataset}
          onClose={() => setSelectedDataset(null)}
        />
      )}

      {showDatasetDialog && (
        <CreateDatasetDialog
          isOpen={showDatasetDialog}
          onClose={() => setShowDatasetDialog(false)}
          onCreate={handleCreateDataset}
          onUpdate={() => {
            console.warn(
              "onUpdate called in DatasetsForm's CreateDatasetDialog instance, but not implemented here.",
            );
          }}
        />
      )}
    </div>
  );
}
