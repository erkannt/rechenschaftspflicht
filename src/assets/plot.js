import * as Plot from "https://cdn.jsdelivr.net/npm/@observablehq/plot@0.6/+esm";

async function drawPlot() {
  const response = await fetch("/events.json");
  if (!response.ok)
    throw new Error(`Failed to load events: ${response.status}`);
  const events = await response.json(); // expects an array of Event objects

  // Create a histogram of events over time (by recordedAt)
  const plot = Plot.rectY(
    events,
    Plot.binX(
      { y: "count" },
      { x: (d) => new Date(d.recordedAt) }, // parse the recordedAt string as a Date
    ),
  ).plot();

  const div = document.querySelector("#myplot");
  div.append(plot);
}

// Run the plotting function
drawPlot().catch(console.error);
