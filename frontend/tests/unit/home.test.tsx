import React from "react";
import { render, screen } from "@testing-library/react";
import Home from "@/app/page";

describe("Home", () => {
  it("renders headline", () => {
    render(<Home />);
    expect(screen.getByText(/VPN MVP Dashboard/i)).toBeInTheDocument();
  });
});
