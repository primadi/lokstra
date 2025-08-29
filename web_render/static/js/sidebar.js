// Sidebar Management - Lokstra Web Framework
// Handles sidebar active menu updates, dropdown states, and sessionStorage state

function setActiveSidebarMenu(menuId) {
  sessionStorage.setItem("activeSidebarMenu", menuId);
}

function restoreSidebarMenuState() {
  const activeMenuId = sessionStorage.getItem("activeSidebarMenu");
  if (activeMenuId) {
    const activeMenu = document.getElementById(activeMenuId);
    if (activeMenu) {
      document.querySelectorAll(".sidebar-nav-link").forEach((link) => {
        link.classList.remove(
          "bg-gray-700",
          "border",
          "border-gray-600",
          "text-white"
        );
        link.classList.add("text-gray-300");
      });
      activeMenu.classList.add(
        "bg-gray-700",
        "border",
        "border-gray-600",
        "text-white"
      );
      activeMenu.classList.remove("text-gray-300");
      const dropdown = activeMenu.closest("li[x-data]");
      if (dropdown && dropdown._x_dataStack) {
        dropdown._x_dataStack[0].menuOpen = true;
      }
    }
  }
}

document.addEventListener("DOMContentLoaded", function () {
  document.querySelectorAll(".sidebar-nav-link").forEach((link) => {
    link.addEventListener("click", function () {
      if (this.id) {
        setActiveSidebarMenu(this.id);
      }
    });
  });
  restoreSidebarMenuState();
});
