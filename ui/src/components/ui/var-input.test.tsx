import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MentionInput, type ContextVariable } from "./var-input";

describe("MentionInput", () => {
  const mockVariables: ContextVariable[] = [
    { display: "User Name", path: "user.name", type: "string" },
    { display: "User Age", path: "user.age", type: "string" },
    { display: "User Email", path: "user.email", type: "string" },
  ];

  it("should render simple input without variables", () => {
    render(
      <TestProvider>
        <MentionInput />
      </TestProvider>,
    );

    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
    expect(input.tagName).toBe("INPUT");
    expect(input).toHaveAttribute("placeholder", "input value");
  });

  it("should render textarea when textarea prop is true", () => {
    render(
      <TestProvider>
        <MentionInput textarea={true} />
      </TestProvider>,
    );

    const textarea = screen.getByRole("textbox");
    expect(textarea).toBeInTheDocument();
    expect(textarea.tagName).toBe("TEXTAREA");
    expect(textarea).toHaveAttribute("placeholder", "input value");
  });

  it("should show dropdown when typing @", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} testAt={true} />
      </TestProvider>,
    );

    expect(screen.getByText("User Name")).toBeInTheDocument();
    expect(screen.getByText("User Age")).toBeInTheDocument();
    expect(screen.getByText("User Email")).toBeInTheDocument();
  });

  it("should insert variable when clicking on dropdown item", async () => {
    const onChange = vi.fn();
    render(
      <TestProvider>
        <MentionInput
          variables={mockVariables}
          onChange={onChange}
          testAt={true}
          value={"@"}
        />
      </TestProvider>,
    );
    await userEvent.click(screen.getByText("User Name"));
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        target: expect.objectContaining({
          value: "{{.user.name}}",
        }),
      }),
    );
  });

  it("should handle keyboard navigation in dropdown", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} testAt={true} value={"@"} />
      </TestProvider>,
    );
    const items = screen.getAllByRole("listitem");
    expect(items[0]).toHaveClass("bg-primary/10");
  });

  it("should format multiple variables correctly", async () => {
    const onChange = vi.fn();
    render(
      <TestProvider>
        <MentionInput
          variables={mockVariables}
          onChange={onChange}
          testAt={true}
          value={"@is cool"}
        />
      </TestProvider>,
    );

    await userEvent.click(screen.getByText("User Name"));

    expect(onChange).toHaveBeenLastCalledWith(
      expect.objectContaining({
        target: expect.objectContaining({
          value: "{{.user.name}} is cool",
        }),
      }),
    );
  });
});
