const SECOND = 1000;
const MINUTE = 60;
const HOUR = MINUTE * 60;
/**
 *
 * @param {string} elementSelector
 */
function initDurationCounter(elementSelector) {
  htmx.on("htmx:load", function load() {
    htmx.off("htmx:load", load);

    const counterElement = document.querySelector(elementSelector);
    if (!counterElement) {
      return;
    }

    const durationStart = counterElement.dataset.start;
    const updateWorkoutDuration = () => {
      let duration = (Date.now() - durationStart) / SECOND;

      const hours = Math.floor(duration / HOUR);
      duration -= hours * HOUR;

      const minutes = Math.floor(duration / MINUTE) % MINUTE;
      duration -= minutes * MINUTE;

      const seconds = Math.floor(duration % MINUTE);

      let formattedDuration = "";
      if (hours > 0) {
        formattedDuration += `${hours.toString().padStart(2, "0")}h `;
      }

      if (minutes > 0) {
        formattedDuration += `${minutes.toString().padStart(2, "0")}m `;
      }
      formattedDuration += `${seconds.toString().padStart(2, "0")}s`;

      counterElement.textContent = formattedDuration;
    };
    const intervalId = setInterval(updateWorkoutDuration, SECOND);

    counterElement.addEventListener("htmx:beforeCleanupElement", function () {
      clearInterval(intervalId);
    });
  });
}
