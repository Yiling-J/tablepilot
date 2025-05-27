import { Card, CardContent } from "@/components/ui/card";
import { PlusCircledIcon } from "@radix-ui/react-icons";

interface AddModelCardProps {
  onClick: () => void;
  disabled?: boolean;
}

export function AddModelCard({ onClick, disabled }: AddModelCardProps) {
  return (
    <Card
      className={`bg-card/60 backdrop-blur-sm shadow-md hover:shadow-accent/50 transition-all duration-300 flex flex-col items-center justify-center text-center p-6 ${
        disabled
          ? "opacity-50 cursor-not-allowed"
          : "cursor-pointer hover:border-primary/70 border border-dashed border-input hover:border-primary/50"
      }`}
      onClick={!disabled ? onClick : undefined}
      role="button"
      aria-label="Add new model"
      tabIndex={disabled ? -1 : 0}
      onKeyDown={(e) => {
        if (!disabled && (e.key === "Enter" || e.key === " ")) {
          e.preventDefault();
          onClick();
        }
      }}
    >
      <CardContent className="flex flex-col items-center justify-center gap-3 p-0">
        <PlusCircledIcon
          className={`h-12 w-12 ${disabled ? "text-muted-foreground/50" : "text-primary/70"}`}
        />
        <p
          className={`text-sm font-medium ${disabled ? "text-muted-foreground/80" : "text-card-foreground"}`}
        >
          Add New Model
        </p>
      </CardContent>
    </Card>
  );
}
