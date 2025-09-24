import React from "react";
import { render, screen } from "@testing-library/react";
import Home from "@/app/page";

describe("Home", () => {
  it("renders headline", () => {
    render(<Home />);
    expect(screen.getByText(/Operations overview/i)).toBeInTheDocument();
  });
});
