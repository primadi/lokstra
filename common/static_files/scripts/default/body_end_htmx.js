document.addEventListener("DOMContentLoaded", function () {
  // Inject LS-Layout header for every htmx request
  document.body.addEventListener("htmx:configRequest", function (evt) {
    var layoutMeta = document.querySelector('meta[name="ls-layout"]')
    var layoutName = layoutMeta ? layoutMeta.content : "base.html"
    evt.detail.headers["LS-Layout"] = layoutName
  })

  // Handle layout changes by full page reload, if layout differs
  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    var layoutMeta = document.querySelector('meta[name="ls-layout"]')
    var currentLayout = layoutMeta?.content ?? "base.html"

    var responseLayout = evt.detail.xhr.getResponseHeader("LS-Layout")
    if (responseLayout && responseLayout !== currentLayout) {
      console.log("Layout changed from", currentLayout, "to", responseLayout)
      evt.preventDefault()
      window.location.href =
        evt.detail.pathInfo.finalRequestPath || window.location.pathname
    } else {
      var xhr = evt.detail.xhr
      var titleMeta = xhr.getResponseHeader("LS-Title") || "Lokstra App"
      var descMeta = xhr.getResponseHeader("LS-Description")

      document.title = titleMeta

      if (descMeta) {
        var currentDescMeta = document.head.querySelector(
          'meta[name="description"]'
        )
        if (currentDescMeta) {
          currentDescMeta.content = descMeta
        } else {
          var newDescMeta = document.createElement("meta")
          newDescMeta.name = "description"
          newDescMeta.content = descMeta
          document.head.appendChild(newDescMeta)
        }
      }
    }
  })
})
