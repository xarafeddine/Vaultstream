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

  // Forms
  loginForm: document.getElementById("login-form"),
  videoDraftForm: document.getElementById("video-draft-form"),

  // Buttons
  btnNewVideo: document.getElementById("btn-new-video"),
  btnEmptyUpload: document.getElementById("btn-empty-upload"),
  btnNewCollection: document.getElementById("btn-new-collection"),
  btnDeleteVideo: document.getElementById("btn-delete-video"),
  fabUpload: document.getElementById("fab-upload"),

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
// Auth Functions
// ============================================
function showAuth() {
  elements.authSection.classList.remove("hidden");
  elements.appSection.classList.add("hidden");
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
      localStorage.setItem("token", data.token);
      showApp();
      await loadVideos();
      showToast("Welcome back!", "success");
    }
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function signup() {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  try {
    const res = await fetch("/api/users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || "Signup failed");
    }

    showToast("Account created! Logging in...", "success");
    await login();
  } catch (error) {
    showToast(error.message, "error");
  }
}

function logout() {
  localStorage.removeItem("token");
  state.videos = [];
  state.currentVideo = null;
  showAuth();
  showToast("Logged out successfully", "success");
}

// ============================================
// Video Functions
// ============================================
async function loadVideos() {
  try {
    const res = await fetch("/api/videos", {
      headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
    });

    if (!res.ok) {
      throw new Error("Failed to load videos");
    }

    state.videos = await res.json();
    renderVideos();
    updateStorageStats();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function createVideoDraft() {
  const title = document.getElementById("video-title").value;
  const description = document.getElementById("video-description").value;

  try {
    const res = await fetch("/api/videos", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("token")}`,
      },
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

    // Open the new video for upload
    openVideoDetail(data);
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function deleteVideo() {
  if (!state.currentVideo) return;

  if (!confirm("Are you sure you want to delete this video?")) return;

  try {
    const res = await fetch(`/api/videos/${state.currentVideo.id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
    });

    if (!res.ok) {
      throw new Error("Failed to delete video");
    }

    closeModal(elements.modalVideoDetail);
    await loadVideos();
    showToast("Video deleted", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function getVideo(videoId) {
  try {
    const res = await fetch(`/api/videos/${videoId}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
    });

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
    const res = await fetch(`/api/thumbnail_upload/${videoId}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
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
      updateVideoDetailModal(video);
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
    const res = await fetch(`/api/video_upload/${videoId}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
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
      updateVideoDetailModal(video);
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
}

function createVideoCard(video) {
  const thumbnailHtml = video.thumbnail_url
    ? `<img src="${video.thumbnail_url}" alt="${escapeHtml(video.title)}" />`
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
  // For now, just show video count
  // In future, this will show actual storage used
  const videoCount = state.videos.length;
  elements.storageText.textContent = `${videoCount} video${
    videoCount !== 1 ? "s" : ""
  } stored`;

  // Placeholder progress (10% per video, max 100%)
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

function updateVideoDetailModal(video) {
  document.getElementById("detail-title").textContent = video.title;
  document.getElementById("detail-description").textContent =
    video.description || "No description";

  // Video player
  const videoPlayer = document.getElementById("detail-video-player");
  if (video.video_url) {
    videoPlayer.src = video.video_url;
    videoPlayer.style.display = "block";
    videoPlayer.load();
  } else {
    videoPlayer.src = "";
    videoPlayer.style.display = "none";
  }

  // Thumbnail
  const thumbnail = document.getElementById("detail-thumbnail");
  if (video.thumbnail_url) {
    thumbnail.src = video.thumbnail_url;
    thumbnail.style.display = "block";
  } else {
    thumbnail.style.display = "none";
  }
}

// ============================================
// Toast Notifications
// ============================================
function showToast(message, type = "info") {
  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.innerHTML = `
    <span>${escapeHtml(message)}</span>
  `;

  elements.toastContainer.appendChild(toast);

  // Auto remove after 4 seconds
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
  elements.loginForm.addEventListener("submit", (e) => {
    e.preventDefault();
    login();
  });

  // Video draft form
  elements.videoDraftForm.addEventListener("submit", (e) => {
    e.preventDefault();
    createVideoDraft();
  });

  // New video buttons
  elements.btnNewVideo.addEventListener("click", () =>
    openModal(elements.modalNewVideo)
  );
  elements.btnEmptyUpload.addEventListener("click", () =>
    openModal(elements.modalNewVideo)
  );
  elements.fabUpload.addEventListener("click", () =>
    openModal(elements.modalNewVideo)
  );

  // Delete video
  elements.btnDeleteVideo.addEventListener("click", deleteVideo);

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
  elements.searchInput.addEventListener(
    "input",
    debounce((e) => {
      state.searchQuery = e.target.value;
      renderVideos();
    }, 300)
  );

  // Sidebar toggle (mobile)
  elements.menuToggle.addEventListener("click", () => {
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
    // Escape to close modals
    if (e.key === "Escape") {
      document.querySelectorAll(".modal-backdrop.active").forEach(closeModal);
    }

    // Ctrl/Cmd + K to focus search
    if ((e.ctrlKey || e.metaKey) && e.key === "k") {
      e.preventDefault();
      elements.searchInput.focus();
    }
  });
}

function setupDropzone(dropzoneId, inputId, onFileSelect) {
  const dropzone = document.getElementById(dropzoneId);
  const input = document.getElementById(inputId);

  if (!dropzone || !input) return;

  // File input change
  input.addEventListener("change", onFileSelect);

  // Drag and drop
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
