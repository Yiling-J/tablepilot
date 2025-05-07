import {
    Provider,
    createProvider,
    deleteProvider,
    getProviders,
    updateProvider,
} from "@/actions";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { DialogTitle } from "@radix-ui/react-dialog";
import { ReloadIcon } from "@radix-ui/react-icons";
import { Edit, PlusCircle, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { ProviderDialog } from "./edit_provider";

interface ProvidersListDialogProps {
  isOpen: boolean;
  setIsOpen: (v: boolean) => void;
}

export function ProvidersListDialog({
  isOpen,
  setIsOpen,
}: ProvidersListDialogProps) {
  const [open, setOpen] = useState(false);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [currentProvider, setCurrentProvider] = useState<Provider | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchProviders = async () => {
    setLoading(true);
    const providers = await getProviders();
    setProviders(providers);
    setLoading(false);
  };

  useEffect(() => {
    fetchProviders();
  }, []);

  const handleCreate = () => {
    setCurrentProvider(null);
    setOpen(true);
  };

  const handleEdit = (provider: Provider) => {
    setCurrentProvider(provider);
    setOpen(true);
  };

  const handleSave = async (provider: Provider) => {
    if (currentProvider) {
      await updateProvider(provider.id.toString(), provider);
    } else {
      await createProvider(provider);
    }
    await fetchProviders();
    setOpen(false);
  };

  const handleDelete = async (provider: Provider) => {
    await deleteProvider(provider.id.toString());
    setProviders(providers.filter((p) => p.id !== provider.id));
  };

  if (loading) {
    return (
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-2xl h-[150px] flex justify-center items-center">
          <DialogTitle />
          <div>
            <ReloadIcon className="animate-spin" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogContent className="max-w-2xl p-4 overflow-hidden flex flex-col max-h-[90vh]">
        <DialogTitle />
        <Card className="ring-0 border-0">
          <div className="flex justify-between items-center">
            <CardHeader>
              <CardTitle>Providers</CardTitle>
              <CardDescription>
                Manage your providers and their models
              </CardDescription>
            </CardHeader>
            <Button onClick={handleCreate} className="mr-2">
              <PlusCircle className="mr-2 h-4 w-4" />
              Add Provider
            </Button>
          </div>
          <CardContent>
            <Table className="w-full table-fixed">
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead className="w-[30%]">Base URL</TableHead>
                  <TableHead>Models</TableHead>
                  <TableHead className="text-center">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {providers.map((provider) => (
                  <TableRow
                    key={
                      provider.id === 0
                        ? `${Math.random().toString(36).substring(2, 9)}`
                        : provider.id
                    }
                  >
                    <TableCell className="font-medium">
                      {provider.name}
                    </TableCell>
                    <TableCell>{provider.type}</TableCell>
                    <TableCell className="w-[30%] truncate">
                      {provider.base_url}
                    </TableCell>
                    <TableCell>{provider.models?.length ?? 0}</TableCell>
                    <TableCell>
                      <div className="flex justify-end space-x-2">
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => handleEdit(provider)}
                          disabled={!provider.editable}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="outline"
                          size="icon"
                          disabled={!provider.editable}
                          onClick={() => handleDelete(provider)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <ProviderDialog
          open={open}
          onOpenChange={setOpen}
          provider={currentProvider}
          onSave={handleSave}
        />
      </DialogContent>
    </Dialog>
  );
}
