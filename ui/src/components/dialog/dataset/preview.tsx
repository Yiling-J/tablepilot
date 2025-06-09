import { previewDataset } from "@/actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { JSONObject } from "@/json";
import React, { useEffect, useState } from "react";

interface PreviewData {
  type: "list" | "csv" | string;
  data?: string[];
  rows?: JSONObject[];
}

interface DatasetPreviewDialogProps {
  datasetId?: string;
  isOpen: boolean;
  onClose: () => void;
}

export const DatasetPreviewDialog: React.FC<DatasetPreviewDialogProps> = ({
  datasetId,
  isOpen,
  onClose,
}) => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<PreviewData | null>(null);

  useEffect(() => {
    if (isOpen && datasetId) {
      setIsLoading(true);
      setError(null);
      setData(null);
      previewDataset(datasetId)
        .then((res) => {
          setData(res);
          console.log("Data loaded successfully", res);
        })
        .catch((err) => {
          console.error("Error fetching dataset preview:", err);
          setError(err.message || "An unexpected error occurred.");
        })
        .finally(() => {
          setIsLoading(false);
        });
    }
  }, [isOpen, datasetId]);

  const handleClose = () => {
    setError(null);
    setData(null);
    setIsLoading(false);
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[425px] md:max-w-[600px] lg:max-w-[800px]">
        <DialogHeader>
          <DialogTitle>Dataset Preview</DialogTitle>
          <DialogDescription>
            Previewing dataset{datasetId ? ` (ID: ${datasetId})` : ""}.{" "}
            {data?.type === "csv" && "First 100 rows."}
          </DialogDescription>
        </DialogHeader>
        <div className="p-4 min-h-[200px]">
          {isLoading && (
            <div>
              <Skeleton className="h-4 w-3/4 mb-2" />
              <Skeleton className="h-4 w-1/2 mb-2" />
              <Skeleton className="h-4 w-5/6" />
            </div>
          )}
          {error && <div className="text-red-500">Error: {error}</div>}
          {!isLoading && !error && data && (
            <>
              {data.type === "list" && Array.isArray(data.data) ? (
                <div className="flex flex-wrap">
                  {(data.data as string[]).map((item, index) => (
                    <Badge
                      key={index}
                      variant="outline"
                      className="mr-2 mb-2 px-3 py-1 bg-blue-100 border-blue-300 text-blue-800"
                    >
                      {item}
                    </Badge>
                  ))}
                </div>
              ) : data.type === "csv" && data.rows ? (
                data.rows.length > 0 ? (
                  <div className="overflow-x-auto max-h-[400px] scrollbar-thumb-rounded-full scrollbar-track-rounded-full scrollbar scrollbar-thumb-stone-500 scrollbar-track-background">
                    {" "}
                    {/* Added max-h for vertical scroll too */}
                    <Table>
                      <TableHeader>
                        <TableRow>
                          {Object.keys(data.rows[0]).map((header) => (
                            <TableHead key={header}>{header}</TableHead>
                          ))}
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {data.rows.map((row, rowIndex) => (
                          <TableRow key={rowIndex}>
                            {Object.keys(data.rows![0]).map(
                              (
                                header, // Use headers from first row for consistent key order
                              ) => (
                                <TableCell key={`${rowIndex}-${header}`}>
                                  {row[header] !== undefined &&
                                  row[header] !== null
                                    ? String(row[header])
                                    : "-"}
                                </TableCell>
                              ),
                            )}
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                ) : (
                  <div>CSV has no rows to display.</div>
                )
              ) : (
                <div>
                  Data loaded. Preview for this data type is not yet implemented
                  or data is empty.
                </div>
              )}
            </>
          )}
          {!isLoading && !error && !data && !datasetId && (
            <div>Please provide a Dataset ID to preview.</div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
