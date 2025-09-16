const main = document.querySelector("main")
if (main) {
  main.style.opacity = "0"
  main.style.transition = "opacity 0.5s"
}

document.addEventListener("DOMContentLoaded", function () {
  // Add some basic styles for demonstration
  const style = document.createElement("style")
  style.textContent = `
    .fade-in {
      animation: fadeIn 0.5s ease-in;
    }
    @keyframes fadeIn {
      from {
        opacity: 0;
        transform: translateY(10px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }`
  document.head.appendChild(style)

  const animation = "fade-in"

  const main = document.querySelector("main")
  if (main) {
    main.style.opacity = "1"
    main.classList.add(animation)
  }

  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    evt.target.classList.remove(animation)
  })

  document.body.addEventListener("htmx:afterSwap", function (evt) {
    evt.target.classList.add(animation)
  })

  document.body.addEventListener("htmx:afterRequest", function (evt) {
    if (evt.detail.successful) {
      const newContent = evt.target.querySelector("[hx-swap-oob]")
      if (newContent) {
        newContent.classList.add(animation)
      }
    }
  })
})
