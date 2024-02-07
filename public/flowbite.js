htmx.on("htmx:load", function load() {
    if(typeof Flowbite === 'undefined') {
        console.warn("Flowbite is not loaded");
    }

    initFlowbite();
})