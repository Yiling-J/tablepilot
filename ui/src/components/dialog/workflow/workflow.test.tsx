import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import WorkflowExecutionDialog from "./workflow";

describe("Workflow Run", () => {
  beforeEach(() => {
    render(
      <TestProvider>
        <WorkflowExecutionDialog
          workflow={{
            id: "w1",
            name: "foo",
            description: "bar",
            variables: [
              { name: "v1", type: "string", default_value: "a", options: [] },
              {
                name: "v2",
                type: "string",
                default_value: "b",
                options: ["b", "c"],
              },
            ],
            steps: [
              { type: "DeleteColumn", payload: { table: "t1", column: "c1" } },
              {
                type: "Generate",
                payload: { table: "t1", count: 1, batch: 1 },
              },
              { type: "DeleteTable", payload: { table: "t1" } },
            ],
          }}
          open={true}
          onOpenChange={() => {}}
        />
      </TestProvider>,
    );
  });
  // case 1: not start yet, verify steps are show properly in left side
  // case 2: click start, user should be asked to input variables first, then variables will be sent to API
  // case 3: Get event from API and run workflow until done, verify left steps updated, and right terminal window also show event message correctly
  // case 4: Get an error event, verify workflow stop, left steps show error icon and right terminal show error message
});
