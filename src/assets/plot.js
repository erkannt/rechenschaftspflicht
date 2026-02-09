import * as Plot from "https://cdn.jsdelivr.net/npm/@observablehq/plot@0.6/+esm";

async function drawPlot() {
  const response = await fetch("/events.json");
  if (!response.ok)
    throw new Error(`Failed to load events: ${response.status}`);
  const events = await response.json(); // expects an array of Event objects

  if (events.length === 0) {
    const div = document.querySelector("#myplot");
    div.innerHTML = "<p>No events with values to plot.</p>";
    return;
  }

  // Create scatter plot with faceting by tag and color by user
  const plot = Plot.plot({
    facet: { data: events, y: "tag" },
    marks: [
      Plot.dot(events, {
        x: (d) => new Date(d.recordedAt),
        y: "valueNum",
        fill: "recordedBy",
        r: 3,
      }),
      Plot.gridX(),
      Plot.gridY(),
    ],
    x: { label: "Time" },
    y: { label: "Value" },
    marginLeft: 80,
    marginBottom: 60,
    height: 800,
    width: 1200,
  });

  const div = document.querySelector("#myplot");
  div.append(plot);
}

// Run the plotting function
drawPlot().catch(console.error);
