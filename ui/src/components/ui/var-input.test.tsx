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

  it("should render basic input without variables", () => {
    render(
      <TestProvider>
        <MentionInput />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
    expect(input).toHaveAttribute("data-placeholder", "Type @ to mention a variable...");
  });

  it("should render textarea when textarea prop is true", () => {
    render(
      <TestProvider>
        <MentionInput textarea={true} />
      </TestProvider>
    );
    
    const textarea = screen.getByRole("textbox");
    expect(textarea).toBeInTheDocument();
    expect(textarea.tagName).toBe("TEXTAREA");
  });

  it("should show dropdown when typing @", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    
    expect(screen.getByText("User Name")).toBeInTheDocument();
    expect(screen.getByText("User Age")).toBeInTheDocument();
    expect(screen.getByText("User Email")).toBeInTheDocument();
  });

  it("should filter variables when typing after @", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@name");
    
    expect(screen.getByText("User Name")).toBeInTheDocument();
    expect(screen.queryByText("User Age")).not.toBeInTheDocument();
    expect(screen.queryByText("User Email")).not.toBeInTheDocument();
  });

  it("should insert variable when clicking on dropdown item", async () => {
    const onChange = vi.fn();
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} onChange={onChange} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    await userEvent.click(screen.getByText("User Name"));
    
    // Check if the variable was inserted with proper formatting
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        target: expect.objectContaining({
          value: "{{.user.name}}"
        })
      })
    );
  });

  it("should handle keyboard navigation in dropdown", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    
    // Press arrow down to select second item
    await userEvent.keyboard("{ArrowDown}");
    const items = screen.getAllByRole("listitem");
    expect(items[1]).toHaveClass("bg-primary/10");
    
    // Press arrow up to select first item
    await userEvent.keyboard("{ArrowUp}");
    expect(items[0]).toHaveClass("bg-primary/10");
  });

  it("should insert variable when pressing Enter on selected item", async () => {
    const onChange = vi.fn();
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} onChange={onChange} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    await userEvent.keyboard("{ArrowDown}");
    await userEvent.keyboard("{Enter}");
    
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        target: expect.objectContaining({
          value: "{{.user.age}}"
        })
      })
    );
  });

  it("should close dropdown when pressing Escape", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    expect(screen.getByText("User Name")).toBeInTheDocument();
    
    await userEvent.keyboard("{Escape}");
    expect(screen.queryByText("User Name")).not.toBeInTheDocument();
  });

  it("should close dropdown when clicking outside", async () => {
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    expect(screen.getByText("User Name")).toBeInTheDocument();
    
    await userEvent.click(document.body);
    expect(screen.queryByText("User Name")).not.toBeInTheDocument();
  });

  it("should format multiple variables correctly", async () => {
    const onChange = vi.fn();
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} onChange={onChange} />
      </TestProvider>
    );
    
    const input = screen.getByRole("textbox");
    await userEvent.type(input, "@");
    await userEvent.click(screen.getByText("User Name"));
    await userEvent.type(input, " is ");
    await userEvent.type(input, "@");
    await userEvent.click(screen.getByText("User Age"));
    await userEvent.type(input, " years old");
    
    expect(onChange).toHaveBeenLastCalledWith(
      expect.objectContaining({
        target: expect.objectContaining({
          value: "{{.user.name}} is {{.user.age}} years old"
        })
      })
    );
  });

  it("should handle multiple lines with variables", async () => {
    const onChange = vi.fn();
    render(
      <TestProvider>
        <MentionInput variables={mockVariables} onChange={onChange} textarea={true} />
      </TestProvider>
    );
    
    const textarea = screen.getByRole("textbox");
    
    // Type first line with a variable
    await userEvent.type(textarea, "Hello ");
    await userEvent.type(textarea, "@");
    await userEvent.click(screen.getByText("User Name"));
    await userEvent.type(textarea, "!");
    
    // Add new line
    await userEvent.keyboard("{Enter}");
    
    // Type second line with a variable
    await userEvent.type(textarea, "Your email is ");
    await userEvent.type(textarea, "@");
    await userEvent.click(screen.getByText("User Email"));
    
    // Add new line
    await userEvent.keyboard("{Enter}");
    
    // Type third line with a variable
    await userEvent.type(textarea, "You are ");
    await userEvent.type(textarea, "@");
    await userEvent.click(screen.getByText("User Age"));
    await userEvent.type(textarea, " years old");
    
    expect(onChange).toHaveBeenLastCalledWith(
      expect.objectContaining({
        target: expect.objectContaining({
          value: "Hello {{.user.name}}!\nYour email is {{.user.email}}\nYou are {{.user.age}} years old"
        })
      })
    );
  });
}); 