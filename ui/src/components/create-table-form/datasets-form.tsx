import { getTableDatasets, createDataset, TableCreateRequest, DatasetInfo, CreateDatasetRequest } from "@/actions";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { DatasetPreviewDialog } from "@/components/dialog/dataset/preview";
import { CreateDatasetDialog } from "@/components/dialog/dataset/dataset";
import { ContextVariable } from "@/components/ui/var-input";
// import { TableInfo } from "@/context/tables"; // Removed
// import { JSONObject } from "@/json"; // Removed as unused
// import { Dataset } from "@/models"; // Replaced with DatasetInfo
import { useEffect, useState } from "react";
import { toast } from "react-hot-toast";

interface DatasetsFormProps {
  form?: TableCreateRequest; // Retaining: passed to CreateDatasetDialog (indirectly via handleCreateDataset)
  table?: string; // Retaining: used for fetching datasets and in CreateDatasetDialog
  variables?: ContextVariable[]; // Retaining: not used directly, but good to keep if DatasetsForm evolves
  // formData: TableCreateRequest; // Removed as unused
  // updateFormData: (data: Partial<TableCreateRequest>) => void; // Removed as unused
  // tables?: TableInfo[]; // Removed as unused
}

export function DatasetsForm({
  table,
  // formData, // Removed
  // updateFormData, // Not used yet, but keeping for future use
  // variables, // Not used yet, but keeping for future use
  // tables, // Not used yet, but keeping for future use
}: DatasetsFormProps) {
  const [datasets, setDatasets] = useState<DatasetInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [selectedDataset, setSelectedDataset] = useState<DatasetInfo | null>(null);
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

  // Signature of datasetData needs to match what CreateDatasetDialog's onSubmit provides
  // which is: { name: string; description: string; type: "list" | "csv"; options?: string[]; files?: File[] }
  const handleCreateDataset = async (datasetData: {
    name: string;
    description: string;
    type: "list" | "csv";
    options?: string[];
    files?: File[];
  }) => {
    if (!table) return; // table (id) is used in the API path for createDataset if it's table-specific, or not if global
    setLoading(true);
    try {
      // Construct CreateDatasetRequest from datasetData
      const requestData: CreateDatasetRequest = {
        name: datasetData.name,
        description: datasetData.description,
        type: datasetData.type,
        data: datasetData.options || [],
        files: datasetData.files || [],
        private: true, // As per original plan
        // table_id: table, // This was causing an error, table_id is not in CreateDatasetRequest
      };
      await createDataset(requestData);
      toast.success("Dataset created successfully!");
      fetchDatasets(); // Refresh the list
    } catch (error) {
      console.error("Failed to create dataset:", error);
      toast.error("Failed to create dataset.");
    } finally {
      setLoading(false);
      setShowDatasetDialog(false);
    }
  };

  if (!table) { // If creating a new table
    // TODO: Allow adding sources/datasets for a new table configuration
    // For now, if it's a new table, this form won't allow adding datasets directly initially.
    // This part of the UI/UX might need further refinement based on product requirements.
    // It's assumed that for new tables, sources/datasets might be added in a different step or manner.
  }

  return (
    <div className="space-y-4">
      {table && (
        <>
          <div className="flex justify-between items-center">
            <h3 className="text-lg font-medium">Datasets</h3>
            <Button onClick={handleAddDataset} variant="outline">Add Dataset</Button>
          </div>
          {loading && <p>Loading datasets...</p>}
          {!loading && datasets.length === 0 && <p>No datasets found for this table.</p>}
          {!loading && datasets.length > 0 && (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {datasets.map((dataset) => (
                  <TableRow key={dataset.id} onClick={() => handleDatasetClick(dataset)} className="cursor-pointer">
                    <TableCell>{dataset.name}</TableCell>
                    <TableCell>{dataset.type}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </>
      )}

      {selectedDataset && (
        <DatasetPreviewDialog
          datasetId={selectedDataset.id}
          isOpen={!!selectedDataset}
          onClose={() => setSelectedDataset(null)}
        />
      )}

      {showDatasetDialog && table && (
        <CreateDatasetDialog
          isOpen={showDatasetDialog}
          onClose={() => setShowDatasetDialog(false)}
          onCreate={handleCreateDataset}
          onUpdate={() => {
            // This dialog instance in DatasetsForm is only for creation.
            // Updates are handled in DatasetListPage or if this form is extended.
            console.warn("onUpdate called in DatasetsForm's CreateDatasetDialog instance, but not implemented here.");
          }}
          // dataset={undefined} // Explicitly not passing dataset for create mode
        />
      )}
    </div>
  );
}
