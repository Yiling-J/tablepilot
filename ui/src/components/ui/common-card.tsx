import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Edit3, Trash2 } from "lucide-react";
import * as React from "react";

interface CommonCardProps {
  name: string;
  children: React.ReactNode;
  onClick?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  className?: string;
}

export function CommonCard({
  name,
  children,
  onClick,
  onEdit,
  onDelete,
  className,
}: CommonCardProps) {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);

  const handleCardClick = (e: React.MouseEvent<HTMLDivElement>) => {
    // Stop propagation if the click is on a button inside the card
    if (e.target instanceof HTMLElement && e.target.closest("button")) {
      return;
    }
    if (onClick) {
      onClick();
    }
  };

  const handleEditClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation(); // Prevent card click when edit button is clicked
    if (onEdit) {
      onEdit();
    }
  };

  const handleDeleteClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation(); // Prevent card click when delete button is clicked
    setIsDeleteDialogOpen(true);
  };

  const confirmDelete = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation();
    if (onDelete) {
      await onDelete();
    }
    setIsDeleteDialogOpen(false);
  };

  const cancelDelete = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation();
    setIsDeleteDialogOpen(false);
  };

  const stopPropagation = (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation();
  }

  return (
    <Card
      className={`h-60 flex flex-col rounded-lg bg-background border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer ${className}`}
    >
      <div
        onClick={handleCardClick}
        className="flex flex-col flex-grow p-4"
      >
        <CardHeader className="p-0 pb-2">
          <CardTitle className="text-lg font-semibold truncate">
            {name}
          </CardTitle>
        </CardHeader>
        <CardContent className="grow mt-2 p-0 overflow-hidden">
          {children}
        </CardContent>
      </div>
      {(onEdit || onDelete) && (
        <CardFooter className="px-4 py-3 border-t border-gray-400/30 flex justify-end gap-2">
          {onEdit && (
            <Button
              variant="ghost"
              size="icon"
              title="Edit"
              onClick={handleEditClick}
            >
              <Edit3 className="h-4 w-4" />
            </Button>
          )}
          {onDelete && (
            <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  title="Delete"
                  className="text-destructive hover:text-destructive"
                  onClick={handleDeleteClick}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent onClick={stopPropagation}>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone. This will permanently delete the item.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel onClick={cancelDelete}>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={confirmDelete}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
        </CardFooter>
      )}
    </Card>
  );
}
