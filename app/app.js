/**
 * Vaultstream - Private Video Vault
 * Main Application JavaScript
 */

// ============================================
// State Management
// ============================================
const state = {
  currentVideo: null,
  videos: [],
  collections: [],
  currentView: "all",
  searchQuery: "",
};

// ============================================
// DOM Elements
// ============================================
const elements = {
  // Sections
  authSection: document.getElementById("auth-section"),
  appSection: document.getElementById("app-section"),

  // Auth Forms
  loginForm: document.getElementById("login-form"),
  signupForm: document.getElementById("signup-form"),
  forgotPasswordForm: document.getElementById("forgot-password-form"),
  resetPasswordForm: document.getElementById("reset-password-form"),

  // Navigation
  sidebar: document.getElementById("sidebar"),
  menuToggle: document.getElementById("menu-toggle"),
  searchInput: document.getElementById("search-input"),

  // Content
  videoGrid: document.getElementById("video-grid"),
  emptyState: document.getElementById("empty-state"),
  pageTitle: document.getElementById("page-title"),
  collectionsList: document.getElementById("collections-list"),

  // Storage
  storageFill: document.getElementById("storage-fill"),
  storageText: document.getElementById("storage-text"),

  // Modals
  modalNewVideo: document.getElementById("modal-new-video"),
  modalVideoDetail: document.getElementById("modal-video-detail"),
  modalEditVideo: document.getElementById("modal-edit-video"),

  // Forms
  videoDraftForm: document.getElementById("video-draft-form"),
  editVideoForm: document.getElementById("edit-video-form"),

  // Buttons
  btnNewVideo: document.getElementById("btn-new-video"),
  btnEmptyUpload: document.getElementById("btn-empty-upload"),
  btnNewCollection: document.getElementById("btn-new-collection"),
  btnVideoActions: document.getElementById("btn-video-actions"),
  fabUpload: document.getElementById("fab-upload"),

  // Dropdown Items
  dropdownDeleteThumbnail: document.getElementById("dropdown-delete-thumbnail"),
  dropdownDeleteVideoFile: document.getElementById(
    "dropdown-delete-video-file"
  ),
  dropdownDeleteAll: document.getElementById("dropdown-delete-all"),

  // Toast
  toastContainer: document.getElementById("toast-container"),
};

// ============================================
// Initialization
// ============================================
document.addEventListener("DOMContentLoaded", async () => {
  const token = localStorage.getItem("token");

  if (token) {
    showApp();
    await loadVideos();
  } else {
    showAuth();
  }

  setupEventListeners();
});

// ============================================
// Auth Form Toggling
// ============================================
function hideAllAuthForms() {
  elements.loginForm?.classList.add("hidden");
  elements.signupForm?.classList.add("hidden");
  elements.forgotPasswordForm?.classList.add("hidden");
  elements.resetPasswordForm?.classList.add("hidden");
}

function showLoginForm() {
  hideAllAuthForms();
  elements.loginForm?.classList.remove("hidden");
}

function showSignupForm() {
  hideAllAuthForms();
  elements.signupForm?.classList.remove("hidden");
}

function showForgotPasswordForm() {
  hideAllAuthForms();
  elements.forgotPasswordForm?.classList.remove("hidden");
}

function showResetPasswordForm() {
  hideAllAuthForms();
  elements.resetPasswordForm?.classList.remove("hidden");
}

// ============================================
// Auth Functions
// ============================================
function showAuth() {
  elements.authSection.classList.remove("hidden");
  elements.appSection.classList.add("hidden");
  showLoginForm();
}

function showApp() {
  elements.authSection.classList.add("hidden");
  elements.appSection.classList.remove("hidden");
}

async function login() {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  try {
    const res = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Login failed");
    }

    if (data.token) {
      // Store both access and refresh tokens
      localStorage.setItem("token", data.token);
      if (data.refresh_token) {
        localStorage.setItem("refresh_token", data.refresh_token);
      }
      showApp();
      await loadVideos();
      showToast("Welcome back!", "success");
    }
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function signup() {
  const fullName = document.getElementById("signup-fullname").value;
  const email = document.getElementById("signup-email").value;
  const password = document.getElementById("signup-password").value;

  if (!fullName || !email || !password) {
    showToast("Please fill in all fields", "error");
    return;
  }

  try {
    const res = await fetch("/api/users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email,
        password,
        full_name: fullName,
      }),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Signup failed");
    }

    showToast("Account created! Logging in...", "success");

    // Copy values to login form and login
    document.getElementById("email").value = email;
    document.getElementById("password").value = password;
    showLoginForm();
    await login();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function forgotPassword() {
  const email = document.getElementById("forgot-email").value;

  if (!email) {
    showToast("Please enter your email", "error");
    return;
  }

  try {
    const res = await fetch("/api/forgot-password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email }),
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Failed to send reset link");
    }

    showToast(data.message, "success");

    // For demo: if token is returned, pre-fill it
    if (data.token) {
      document.getElementById("reset-token").value = data.token;
      showResetPasswordForm();
    }
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function resetPassword() {
  const token = document.getElementById("reset-token").value;
  const newPassword = document.getElementById("reset-new-password").value;

  if (!token || !newPassword) {
    showToast("Please fill in all fields", "error");
    return;
  }

  try {
    const res = await fetch("/api/reset-password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token, new_password: newPassword }),
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Failed to reset password");
    }

    showToast(data.message, "success");
    showLoginForm();
  } catch (error) {
    showToast(error.message, "error");
  }
}

function logout() {
  localStorage.removeItem("token");
  localStorage.removeItem("refresh_token");
  state.videos = [];
  state.currentVideo = null;
  showAuth();
  showToast("Logged out successfully", "success");
}

// ============================================
// Token Refresh
// ============================================
async function refreshAccessToken() {
  const refreshToken = localStorage.getItem("refresh_token");
  if (!refreshToken) {
    return false;
  }

  try {
    const res = await fetch("/api/refresh", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${refreshToken}`,
      },
    });

    if (!res.ok) {
      throw new Error("Token refresh failed");
    }

    const data = await res.json();
    if (data.token) {
      localStorage.setItem("token", data.token);
      return true;
    }
    return false;
  } catch (error) {
    console.error("Token refresh failed:", error);
    return false;
  }
}

// Wrapper for authenticated API calls with auto-refresh
async function authFetch(url, options = {}) {
  options.headers = options.headers || {};
  options.headers["Authorization"] = `Bearer ${localStorage.getItem("token")}`;

  let res = await fetch(url, options);

  // If unauthorized, try to refresh token
  if (res.status === 401) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      // Retry with new token
      options.headers["Authorization"] = `Bearer ${localStorage.getItem(
        "token"
      )}`;
      res = await fetch(url, options);
    } else {
      // Refresh failed, logout
      logout();
      showToast("Session expired. Please log in again.", "warning");
      throw new Error("Session expired");
    }
  }

  return res;
}

// ============================================
// Video Functions
// ============================================
async function loadVideos() {
  try {
    const res = await authFetch("/api/videos");

    if (!res.ok) {
      throw new Error("Failed to load videos");
    }

    state.videos = await res.json();
    renderVideos();
    updateStorageStats();
  } catch (error) {
    if (error.message !== "Session expired") {
      showToast(error.message, "error");
    }
  }
}

async function createVideoDraft() {
  const title = document.getElementById("video-title").value;
  const description = document.getElementById("video-description").value;

  try {
    const res = await authFetch("/api/videos", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, description }),
    });

    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || "Failed to create video");
    }

    closeModal(elements.modalNewVideo);
    document.getElementById("video-title").value = "";
    document.getElementById("video-description").value = "";

    await loadVideos();
    showToast("Video draft created!", "success");

    // Open the new video edit details immediately
    state.currentVideo = data;
    openEditModal();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function updateVideoDetails() {
  if (!state.currentVideo) return;

  const title = document.getElementById("edit-video-title").value;
  const description = document.getElementById("edit-video-description").value;

  try {
    const res = await authFetch(`/api/videos/${state.currentVideo.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, description }),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Failed to update video");
    }

    const updatedVideo = await res.json();
    state.currentVideo = updatedVideo;

    // Update local list
    const index = state.videos.findIndex((v) => v.id === updatedVideo.id);
    if (index !== -1) {
      state.videos[index] = updatedVideo;
    }

    closeModal(elements.modalEditVideo);
    renderVideos();

    // If detail modal is open, update it
    if (elements.modalVideoDetail.classList.contains("active")) {
      updateVideoDetailModal(updatedVideo);
    } else {
      // Open detail modal if not open (edit flow finished)
      openVideoDetail(updatedVideo);
    }

    showToast("Video updated successfully", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function deleteVideo() {
  if (!state.currentVideo) return;

  if (!confirm("Are you sure you want to delete this video and all its files?"))
    return;

  try {
    const res = await authFetch(`/api/videos/${state.currentVideo.id}`, {
      method: "DELETE",
    });

    if (!res.ok) {
      throw new Error("Failed to delete video");
    }

    closeModal(elements.modalVideoDetail);
    closeModal(elements.modalEditVideo);
    await loadVideos();
    showToast("Video deleted", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function deleteThumbnailOnly() {
  if (!state.currentVideo) return;

  if (!confirm("Delete only the thumbnail?")) return;

  try {
    const res = await authFetch(
      `/api/videos/${state.currentVideo.id}/thumbnail`,
      {
        method: "DELETE",
      }
    );

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Failed to delete thumbnail");
    }

    showToast("Thumbnail deleted", "success");

    // Refresh video data
    const video = await getVideo(state.currentVideo.id);
    if (video) {
      state.currentVideo = video;
      // Update both modals if open
      if (elements.modalVideoDetail.classList.contains("active"))
        updateVideoDetailModal(video);
      // Re-populate edit modal if open
      if (elements.modalEditVideo.classList.contains("active")) {
        // No special update needed except global state?
        // Actually we might want to refresh lists, but let's just refresh video list
      }
      await loadVideos();
    }
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function deleteVideoFileOnly() {
  if (!state.currentVideo) return;

  if (!confirm("Delete only the video file?")) return;

  try {
    const res = await authFetch(
      `/api/videos/${state.currentVideo.id}/video-file`,
      {
        method: "DELETE",
      }
    );

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Failed to delete video file");
    }

    showToast("Video file deleted", "success");

    // Refresh video data
    const video = await getVideo(state.currentVideo.id);
    if (video) {
      state.currentVideo = video;
      if (elements.modalVideoDetail.classList.contains("active"))
        updateVideoDetailModal(video);
      await loadVideos();
    }
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function getVideo(videoId) {
  try {
    const res = await authFetch(`/api/videos/${videoId}`);

    if (!res.ok) {
      throw new Error("Failed to get video");
    }

    return await res.json();
  } catch (error) {
    showToast(error.message, "error");
    return null;
  }
}

// ============================================
// Upload Functions
// ============================================
async function uploadThumbnail(videoId) {
  const fileInput = document.getElementById("thumbnail-input");
  const file = fileInput.files[0];
  if (!file) return;

  const formData = new FormData();
  formData.append("thumbnail", file);

  const progressContainer = document.getElementById("thumbnail-progress");
  progressContainer.classList.remove("hidden");

  try {
    const res = await authFetch(`/api/thumbnail_upload/${videoId}`, {
      method: "POST",
      body: formData,
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Upload failed");
    }

    showToast("Thumbnail uploaded!", "success");

    // Refresh video data
    const video = await getVideo(videoId);
    if (video) {
      state.currentVideo = video;
      // Note: we are usually in edit modal when uploading
      await loadVideos();
    }
  } catch (error) {
    showToast(error.message, "error");
  } finally {
    progressContainer.classList.add("hidden");
    fileInput.value = "";
  }
}

async function uploadVideo(videoId) {
  const fileInput = document.getElementById("video-input");
  const file = fileInput.files[0];
  if (!file) return;

  const formData = new FormData();
  formData.append("video", file);

  const progressContainer = document.getElementById("video-progress");
  progressContainer.classList.remove("hidden");

  try {
    const res = await authFetch(`/api/video_upload/${videoId}`, {
      method: "POST",
      body: formData,
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Upload failed");
    }

    showToast("Video uploaded!", "success");

    // Refresh video data
    const video = await getVideo(videoId);
    if (video) {
      state.currentVideo = video;
      await loadVideos();
    }
  } catch (error) {
    showToast(error.message, "error");
  } finally {
    progressContainer.classList.add("hidden");
    fileInput.value = "";
  }
}

// ============================================
// Expired URL Handling
// ============================================
function handleMediaError(element, type) {
  const parent = element.parentElement;

  // Create expired overlay
  const overlay = document.createElement("div");
  overlay.className = "expired-overlay";
  overlay.innerHTML = `
    <div class="expired-content">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <polyline points="12 6 12 12 16 14"></polyline>
      </svg>
      <p>${type === "video" ? "Video" : "Image"} URL expired</p>
      <button class="btn btn-primary btn-sm" onclick="refreshCurrentVideo()">
        Refresh
      </button>
    </div>
  `;

  element.style.display = "none";
  parent.appendChild(overlay);
}

async function refreshCurrentVideo() {
  if (!state.currentVideo) return;

  showToast("Refreshing...", "info");
  const video = await getVideo(state.currentVideo.id);
  if (video) {
    state.currentVideo = video;
    updateVideoDetailModal(video);
    showToast("Refreshed!", "success");
  }
}

// ============================================
// Rendering Functions
// ============================================
function renderVideos() {
  const filteredVideos = filterVideos();

  if (filteredVideos.length === 0) {
    elements.videoGrid.classList.add("hidden");
    elements.emptyState.classList.remove("hidden");
    return;
  }

  elements.videoGrid.classList.remove("hidden");
  elements.emptyState.classList.add("hidden");

  elements.videoGrid.innerHTML = filteredVideos
    .map((video) => createVideoCard(video))
    .join("");

  // Add click handlers
  elements.videoGrid.querySelectorAll(".video-card").forEach((card) => {
    card.addEventListener("click", () => {
      const videoId = card.dataset.videoId;
      const video = state.videos.find((v) => v.id === videoId);
      if (video) openVideoDetail(video);
    });
  });

  // Add error handlers for thumbnails
  elements.videoGrid
    .querySelectorAll(".video-card-thumbnail img")
    .forEach((img) => {
      img.addEventListener("error", () => {
        img.style.display = "none";
        img.parentElement.innerHTML = `
        <div class="video-card-no-thumbnail">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"></rect>
            <line x1="7" y1="2" x2="7" y2="22"></line>
            <line x1="17" y1="2" x2="17" y2="22"></line>
            <line x1="2" y1="12" x2="22" y2="12"></line>
          </svg>
        </div>
      `;
      });
    });
}

function createVideoCard(video) {
  const thumbnailHtml = video.thumbnail_url
    ? `<img src="${video.thumbnail_url}" alt="${escapeHtml(
        video.title
      )}" onerror="this.style.display='none'" />`
    : `<div class="video-card-no-thumbnail">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"></rect>
          <line x1="7" y1="2" x2="7" y2="22"></line>
          <line x1="17" y1="2" x2="17" y2="22"></line>
          <line x1="2" y1="12" x2="22" y2="12"></line>
        </svg>
       </div>`;

  const date = new Date(video.created_at).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  return `
    <div class="video-card" data-video-id="${video.id}">
      <div class="video-card-thumbnail">
        ${thumbnailHtml}
        <div class="video-card-overlay">
          <div class="video-card-play">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="white">
              <polygon points="5 3 19 12 5 21 5 3"></polygon>
            </svg>
          </div>
        </div>
      </div>
      <div class="video-card-body">
        <h3 class="video-card-title">${escapeHtml(video.title)}</h3>
        <div class="video-card-meta">
          <span>${date}</span>
          ${video.video_url ? "<span>• Ready</span>" : "<span>• Draft</span>"}
        </div>
      </div>
    </div>
  `;
}

function filterVideos() {
  let videos = [...state.videos];

  // Filter by search
  if (state.searchQuery) {
    const query = state.searchQuery.toLowerCase();
    videos = videos.filter(
      (v) =>
        v.title.toLowerCase().includes(query) ||
        (v.description && v.description.toLowerCase().includes(query))
    );
  }

  // Sort by date (newest first)
  videos.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  return videos;
}

function updateStorageStats() {
  const videoCount = state.videos.length;
  elements.storageText.textContent = `${videoCount} video${
    videoCount !== 1 ? "s" : ""
  } stored`;

  const percentage = Math.min(videoCount * 10, 100);
  elements.storageFill.style.width = `${percentage}%`;
}

// ============================================
// Modal Functions
// ============================================
function openModal(modal) {
  modal.classList.add("active");
  document.body.style.overflow = "hidden";
}

function closeModal(modal) {
  modal.classList.remove("active");
  document.body.style.overflow = "";
}

function openVideoDetail(video) {
  state.currentVideo = video;
  updateVideoDetailModal(video);
  openModal(elements.modalVideoDetail);
}

function openEditModal() {
  if (!state.currentVideo) return;

  // Close detail modal if open
  closeModal(elements.modalVideoDetail);

  // Populate form
  document.getElementById("edit-video-title").value = state.currentVideo.title;
  document.getElementById("edit-video-description").value =
    state.currentVideo.description || "";

  // Open edit modal
  openModal(elements.modalEditVideo);
}

function updateVideoDetailModal(video) {
  document.getElementById("detail-title").textContent = video.title;
  document.getElementById("detail-description").textContent =
    video.description || "No description";

  // Video player with error handling
  const videoPlayer = document.getElementById("detail-video-player");
  const videoContainer = videoPlayer.parentElement;

  // Remove any existing overlay
  const existingOverlay = videoContainer.querySelector(".expired-overlay");
  if (existingOverlay) existingOverlay.remove();

  if (video.video_url) {
    videoPlayer.src = video.video_url;
    videoPlayer.style.display = "block";
    videoPlayer.onerror = () => handleMediaError(videoPlayer, "video");
    videoPlayer.load();
  } else {
    videoPlayer.src = "";
    videoPlayer.style.display = "none";
  }

  // Thumbnail with error handling
  const thumbnail = document.getElementById("detail-thumbnail");
  const thumbContainer = thumbnail.parentElement;
  const noThumbText = document.getElementById("no-thumbnail-text");

  // Remove any existing overlay
  const existingThumbOverlay = thumbContainer.querySelector(".expired-overlay");
  if (existingThumbOverlay) existingThumbOverlay.remove();

  if (video.thumbnail_url) {
    thumbnail.src = video.thumbnail_url;
    thumbnail.style.display = "block";
    noThumbText.style.display = "none";
    thumbnail.onerror = () => handleMediaError(thumbnail, "image");
  } else {
    thumbnail.style.display = "none";
    noThumbText.style.display = "block";
  }

  // Update actions visibility in dropdown
  updateActionsDropdown(video);
}

function updateActionsDropdown(video) {
  const btnThumb = elements.dropdownDeleteThumbnail;
  const btnVideo = elements.dropdownDeleteVideoFile;

  if (btnThumb) {
    btnThumb.style.display = video.thumbnail_url ? "flex" : "none";
  }
  if (btnVideo) {
    btnVideo.style.display = video.video_url ? "flex" : "none";
  }
}

function toggleVideoActionsDropdown() {
  const dropdown = document.getElementById("video-actions-dropdown");
  dropdown.classList.toggle("open");
}

function closeDropdowns() {
  document
    .querySelectorAll(".dropdown.open")
    .forEach((d) => d.classList.remove("open"));
}

// ============================================
// Toast Notifications
// ============================================
function showToast(message, type = "info") {
  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.innerHTML = `<span>${escapeHtml(message)}</span>`;

  elements.toastContainer.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = "0";
    toast.style.transform = "translateY(20px)";
    setTimeout(() => toast.remove(), 300);
  }, 4000);
}

// ============================================
// Event Listeners
// ============================================
function setupEventListeners() {
  // Login form
  elements.loginForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    login();
  });

  // Signup form
  elements.signupForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    signup();
  });

  // Forgot password form
  elements.forgotPasswordForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    forgotPassword();
  });

  // Reset password form
  elements.resetPasswordForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    resetPassword();
  });

  // Video draft form
  elements.videoDraftForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    createVideoDraft();
  });

  // Edit video form
  elements.editVideoForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    updateVideoDetails();
  });

  // New video buttons
  elements.btnNewVideo?.addEventListener("click", () =>
    openModal(elements.modalNewVideo)
  );
  elements.btnEmptyUpload?.addEventListener("click", () =>
    openModal(elements.modalNewVideo)
  );
  elements.fabUpload?.addEventListener("click", () =>
    openModal(elements.modalNewVideo)
  );

  // Toggle actions dropdown
  elements.btnVideoActions?.addEventListener("click", (e) => {
    e.stopPropagation();
    toggleVideoActionsDropdown();
  });

  // Close dropdowns when clicking outside
  document.addEventListener("click", () => closeDropdowns());

  // Dropdown actions
  elements.dropdownDeleteAll?.addEventListener("click", deleteVideo);
  elements.dropdownDeleteThumbnail?.addEventListener(
    "click",
    deleteThumbnailOnly
  );
  elements.dropdownDeleteVideoFile?.addEventListener(
    "click",
    deleteVideoFileOnly
  );

  // Close modal buttons
  document.querySelectorAll("[data-close-modal]").forEach((btn) => {
    btn.addEventListener("click", (e) => {
      const modal = e.target.closest(".modal-backdrop");
      if (modal) closeModal(modal);
    });
  });

  // Close modal on backdrop click
  document.querySelectorAll(".modal-backdrop").forEach((modal) => {
    modal.addEventListener("click", (e) => {
      if (e.target === modal) closeModal(modal);
    });
  });

  // Search
  elements.searchInput?.addEventListener(
    "input",
    debounce((e) => {
      state.searchQuery = e.target.value;
      renderVideos();
    }, 300)
  );

  // Sidebar toggle (mobile)
  elements.menuToggle?.addEventListener("click", () => {
    elements.sidebar.classList.toggle("open");
  });

  // Sidebar navigation
  document.querySelectorAll(".sidebar-nav-link[data-view]").forEach((link) => {
    link.addEventListener("click", () => {
      document
        .querySelectorAll(".sidebar-nav-link")
        .forEach((l) => l.classList.remove("active"));
      link.classList.add("active");
      state.currentView = link.dataset.view;
      elements.pageTitle.textContent = link.textContent.trim();
      renderVideos();
    });
  });

  // File uploads with drag and drop
  setupDropzone("thumbnail-dropzone", "thumbnail-input", () => {
    if (state.currentVideo) uploadThumbnail(state.currentVideo.id);
  });

  setupDropzone("video-dropzone", "video-input", () => {
    if (state.currentVideo) uploadVideo(state.currentVideo.id);
  });

  // Keyboard shortcuts
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") {
      document.querySelectorAll(".modal-backdrop.active").forEach(closeModal);
      closeDropdowns();
    }

    if ((e.ctrlKey || e.metaKey) && e.key === "k") {
      e.preventDefault();
      elements.searchInput?.focus();
    }
  });
}

function setupDropzone(dropzoneId, inputId, onFileSelect) {
  const dropzone = document.getElementById(dropzoneId);
  const input = document.getElementById(inputId);

  if (!dropzone || !input) return;

  input.addEventListener("change", onFileSelect);

  ["dragenter", "dragover"].forEach((event) => {
    dropzone.addEventListener(event, (e) => {
      e.preventDefault();
      dropzone.classList.add("dragover");
    });
  });

  ["dragleave", "drop"].forEach((event) => {
    dropzone.addEventListener(event, (e) => {
      e.preventDefault();
      dropzone.classList.remove("dragover");
    });
  });

  dropzone.addEventListener("drop", (e) => {
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      input.files = files;
      onFileSelect();
    }
  });
}

// ============================================
// Utility Functions
// ============================================
function escapeHtml(text) {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

function formatBytes(bytes) {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

// Make functions available globally for onclick handlers
window.logout = logout;
window.signup = signup;
window.showLoginForm = showLoginForm;
window.showSignupForm = showSignupForm;
window.showForgotPasswordForm = showForgotPasswordForm;
window.showResetPasswordForm = showResetPasswordForm;
window.refreshCurrentVideo = refreshCurrentVideo;
window.deleteThumbnailOnly = deleteThumbnailOnly;
window.deleteVideoFileOnly = deleteVideoFileOnly;
window.openEditModal = openEditModal;
window.toggleVideoActionsDropdown = toggleVideoActionsDropdown;
