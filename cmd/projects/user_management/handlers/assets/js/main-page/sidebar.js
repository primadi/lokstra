// Sidebar Management - User Management Application
// Handles sidebar active menu updates, dropdown states, and sessionStorage state

// Simpan id menu yang aktif ke sessionStorage
function setActiveSidebarMenu(menuId) {
  sessionStorage.setItem("activeSidebarMenu", menuId);
}

// Restore state menu sidebar dari sessionStorage
function restoreSidebarMenuState() {
  const activeMenuId = sessionStorage.getItem("activeSidebarMenu");
  if (activeMenuId) {
    const activeMenu = document.getElementById(activeMenuId);
    if (activeMenu) {
      // Remove active styles dari semua menu
      document.querySelectorAll(".sidebar-nav-link").forEach((link) => {
        link.classList.remove(
          "bg-gray-700",
          "border",
          "border-gray-600",
          "text-white"
        );
        link.classList.add("text-gray-300");
      });
      // Tambahkan active style ke menu yang restore
      activeMenu.classList.add(
        "bg-gray-700",
        "border",
        "border-gray-600",
        "text-white"
      );
      activeMenu.classList.remove("text-gray-300");
      // Handle dropdown jika menu ada di dalam dropdown
      const dropdown = activeMenu.closest("li[x-data]");
      if (dropdown && dropdown._x_dataStack) {
        dropdown._x_dataStack[0].menuOpen = true;
      }
    }
  }
}

// Event listener untuk klik menu sidebar
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
