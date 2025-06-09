import "@testing-library/jest-dom";
import {
    fireEvent,
    render,
    screen,
    waitFor,
    within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { CommonCard } from "./common-card";

// Mock Lucide icons
vi.mock("lucide-react", async () => {
  const actual =
    await vi.importActual<typeof import("lucide-react")>("lucide-react");
  return {
    ...actual,
    Edit3: () => <div data-testid="edit-icon" />,
    Trash2: () => <div data-testid="delete-icon" />,
  };
});

describe("CommonCard", () => {
  const defaultProps = {
    name: "Test Card",
    children: <p>Test content</p>,
    onClick: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn(),
  };

  afterEach(() => {
    vi.clearAllMocks();
  });

  test("renders with required props", () => {
    render(<CommonCard name="Test Card" children={<p>Test content</p>} />);
    expect(screen.getByText("Test Card")).toBeInTheDocument();
    expect(screen.getByText("Test content")).toBeInTheDocument();
  });

  test("calls onClick when card is clicked", () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(
      screen
        .getByText("Test Card")
        .closest(".flex-col.flex-grow.p-4")! as HTMLElement,
    );
    expect(defaultProps.onClick).toHaveBeenCalledTimes(1);
  });

  test("renders edit and delete buttons when onEdit and onDelete are provided", () => {
    render(<CommonCard {...defaultProps} />);
    expect(screen.getByTitle("Edit")).toBeInTheDocument();
    expect(screen.getByTitle("Delete")).toBeInTheDocument();
  });

  test("does not render edit button if onEdit is not provided", () => {
    render(
      <CommonCard
        name="Test Card"
        children={<p>Test content</p>}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.queryByTitle("Edit")).not.toBeInTheDocument();
  });

  test("does not render delete button if onDelete is not provided", () => {
    render(
      <CommonCard
        name="Test Card"
        children={<p>Test content</p>}
        onEdit={vi.fn()}
      />,
    );
    expect(screen.queryByTitle("Delete")).not.toBeInTheDocument();
  });

  test("calls onEdit when edit button is clicked", () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(screen.getByTitle("Edit") as HTMLElement);
    expect(defaultProps.onEdit).toHaveBeenCalledTimes(1);
    expect(defaultProps.onClick).not.toHaveBeenCalled(); // Ensure card click is not also triggered
  });

  test("opens delete confirmation dialog when delete button is clicked", () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(screen.getByTitle("Delete") as HTMLElement);
    expect(screen.getByText("Are you sure?")).toBeInTheDocument();
    expect(
      screen.getByText(
        "This action cannot be undone. This will permanently delete the item.",
      ),
    ).toBeInTheDocument();
  });

  test("calls onDelete when delete is confirmed", async () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(screen.getByTitle("Delete") as HTMLElement); // Open dialog
    // Wait for the dialog to be fully open and interactive if necessary
    await waitFor(() =>
      expect(
        screen.getByText("Delete", { selector: "button.bg-destructive" }),
      ).toBeVisible(),
    );
    fireEvent.click(
      screen.getByText("Delete", {
        selector: "button.bg-destructive",
      }) as HTMLElement,
    ); // Click the confirm delete button

    expect(defaultProps.onDelete).toHaveBeenCalledTimes(1);
    // Dialog should close after deletion
    await waitFor(() => {
      expect(screen.queryByText("Are you sure?")).not.toBeInTheDocument();
    });
  });

  test("does not call onDelete when delete is cancelled", async () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(screen.getByTitle("Delete") as HTMLElement); // Open dialog
    // Wait for the dialog to be fully open and interactive if necessary
    await waitFor(() => expect(screen.getByText("Cancel")).toBeVisible());
    fireEvent.click(screen.getByText("Cancel") as HTMLElement); // Click the cancel button

    expect(defaultProps.onDelete).not.toHaveBeenCalled();
    // Dialog should close
    await waitFor(() => {
      expect(screen.queryByText("Are you sure?")).not.toBeInTheDocument();
    });
  });

  test("card onClick is not triggered when edit button is clicked", () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(screen.getByTitle("Edit") as HTMLElement);
    expect(defaultProps.onEdit).toHaveBeenCalledTimes(1);
    expect(defaultProps.onClick).not.toHaveBeenCalled();
  });

  test("card onClick is not triggered when delete button is clicked (and dialog opens)", () => {
    render(<CommonCard {...defaultProps} />);
    fireEvent.click(screen.getByTitle("Delete") as HTMLElement);
    // Dialog opens, onClick for card should not have been called
    expect(defaultProps.onClick).not.toHaveBeenCalled();
  });

  test("card onClick is not triggered when an action inside AlertDialog is clicked", async () => {
    // Part 1: Test delete confirmation
    const onClickMock1 = vi.fn();
    const { unmount: unmount1 } = render(
      <CommonCard {...defaultProps} onClick={onClickMock1} />,
    );
    // The card's delete button should be unique enough with getByTitle before dialog opens
    fireEvent.click(screen.getByTitle("Delete") as HTMLElement);

    // Dialog elements are often portalled to document.body, so screen queries are more robust
    await waitFor(() =>
      expect(
        screen.getByText("Delete", { selector: "button.bg-destructive" }),
      ).toBeVisible(),
    );
    fireEvent.click(
      screen.getByText("Delete", {
        selector: "button.bg-destructive",
      }) as HTMLElement,
    );
    expect(onClickMock1).not.toHaveBeenCalled();
    unmount1(); // Cleanup the first render

    // Part 2: Test cancel
    const onClickMock2 = vi.fn();
    // Render a card with a different name to ensure we can target its specific delete trigger
    const { unmount: unmount2 } = render(
      <CommonCard
        {...defaultProps}
        name="Test Card 2"
        onClick={onClickMock2}
      />,
    );
    // Find the delete trigger for "Test Card 2".
    // We need to find the card first, then the delete button within it to be specific.
    const card2 = screen.getByText("Test Card 2").closest('div[class*="h-60"]'); // Find the card root
    if (!card2) throw new Error("Could not find Test Card 2");
    // Use getByRole for more specific HTMLElement typing and assert
    const deleteButton = within(card2 as HTMLElement).getByRole("button", {
      name: /delete/i,
    });
    await userEvent.click(deleteButton as HTMLElement);

    await waitFor(() => expect(screen.getByText("Cancel")).toBeVisible());
    // Assuming Cancel is a button, target by role and assert
    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await userEvent.click(cancelButton as HTMLElement);
    expect(onClickMock2).not.toHaveBeenCalled();
    unmount2(); // Cleanup the second render
  });

  describe("CommonCard Badge functionality", () => {
    it("renders the badge with correct text when badgeText is provided", () => {
      const badgeText = "Test Badge";
      render(
        <CommonCard name="Test Card" badgeText={badgeText}>
          <p>Test Content</p>
        </CommonCard>,
      );
      const badgeElement = screen.getByText(badgeText);
      expect(badgeElement).toBeInTheDocument();
    });

    it("does not render the badge when badgeText is not provided", () => {
      render(
        <CommonCard name="Test Card">
          <p>Test Content</p>
        </CommonCard>,
      );
      const potentialBadgeText = "Dataset"; // Example text that a badge might have
      expect(screen.queryByText(potentialBadgeText)).not.toBeInTheDocument();
    });

    it("does not render the badge when badgeText is an empty string (as it's falsy)", () => {
      const badgeText = "";
      const { container } = render(
        <CommonCard name="Test Card" badgeText={badgeText}>
          <p>Test Content</p>
        </CommonCard>,
      );
      const badgeElement = container.querySelector(".absolute.top-2.right-2");
      expect(badgeElement).not.toBeInTheDocument();
    });
  });
});
