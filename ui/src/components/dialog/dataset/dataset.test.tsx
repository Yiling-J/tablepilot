import { TestProvider } from "@/test/helpers/test-provider";
import { render, screen } from "@testing-library/react";
import { useNavigate } from "react-router-dom";
import { vi } from "vitest";
import { CreateDatasetDialog } from "./dataset";

describe("Create Dataset", () => {
  beforeEach(async () => {
    vi.mock("react-router-dom");
    vi.mocked(useNavigate).mockReturnValue(vi.fn());
    vi.mock("@/actions");
    render(
      <TestProvider>
        <CreateDatasetDialog
          isOpen={true}
          onClose={() => {}}
          onCreate={() => {}}
        />
      </TestProvider>,
    );
    await screen.findByText("Steps");
  });

  it("should create list type dataset", async () => {});
  it("should create csv type dataset", async () => {});
});
