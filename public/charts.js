const activeCharts = new Map();

/**
 * @param {() => any} getOptions
 * @param {string} elementSelector
 * @param {((chart: any) => any) | undefined } onLoad
 * @param {(() => any) | undefined } onDestroy
 */
function loadChart(getOptions, elementSelector, onLoad, onDestroy) {
  htmx.on("htmx:load", function init() {
    htmx.off("htmx:load", init);

    if (typeof ApexCharts === "undefined") {
      console.error(`ApexCharts not loaded`);
      return;
    }

    const chartElement = document.querySelector(elementSelector);
    if (!chartElement) {
      console.error(`No element found "${elementSelector}"`);
      return;
    }

    try {
      if (activeCharts.has(elementSelector)) {
        const chart = activeCharts.get(elementSelector);
        chart.updateOptions(getOptions(), true, true);
      } else {
        const chart = new ApexCharts(chartElement, getOptions());
        chart.render();

        activeCharts.set(elementSelector, chart);

        if (onLoad) {
          onLoad(chart);
        }
        chartElement.addEventListener("htmx:beforeCleanupElement", function () {
          console.log("Destroy:", elementSelector);
          chart.destroy();
          activeCharts.delete(elementSelector)
          if (onDestroy) {
            onDestroy();
          }
        });
      }
    } catch (error) {
        console.error('Load chart error: ', error)
    }
  });
}
