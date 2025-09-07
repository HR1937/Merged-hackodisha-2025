// Global variables
let currentUser = null;
let currentTab = "feed";
let currentSection = "all";
// API Configuration
const API_BASE = "";
const CLOUD_NAME = "dsavieizd";
const UPLOAD_PRESET = "demo_unsigned";
// Initialize app
document.addEventListener("DOMContentLoaded", () => {
  console.log("[v0] App initializing...");
  initializeApp();
});
function initializeApp() {
  // Hide loading screen after 1 second
  setTimeout(() => {
    document.getElementById("loading").style.display = "none";
    console.log("[v0] Loading screen hidden");
  }, 1000);
  // Check if user is logged in
  const token = localStorage.getItem("token");
  if (token) {
    console.log("[v0] Token found, attempting auto-login");
    showMainApp();
    loadUserProfile();
    loadFeed();
  } else {
    console.log("[v0] No token found, showing homepage");
    showHomepage();
  }
  // Setup event listeners
  setupEventListeners();
}

function setupEventListeners() {
  console.log("[v0] Setting up event listeners");
  // Homepage buttons
  document.getElementById("loginBtn")?.addEventListener("click", showAuthModal);
  document
    .getElementById("getStartedBtn")
    ?.addEventListener("click", showAuthModal);
  document
    .getElementById("learnMoreBtn")
    ?.addEventListener("click", scrollToFeatures);
  // New features
  document.getElementById("neighbourHelpBtn")?.addEventListener("click", triggerNeighbourHelp);
  document.getElementById("chatbotBtn")?.addEventListener("click", () => {
    document.getElementById("chatbotAudioInput").click();
  });
  document.getElementById("chatbotAudioInput")?.addEventListener("change", handleChatbotAudioUpload);
  // Auth modal
  document
    .getElementById("closeAuth")
    ?.addEventListener("click", hideAuthModal);
  document
    .getElementById("showSignup")
    ?.addEventListener("click", showSignupForm);
  document
    .getElementById("showLogin")
    ?.addEventListener("click", showLoginForm);
  document
    .getElementById("loginSubmit")
    ?.addEventListener("click", handleLogin);
  document
    .getElementById("signupSubmit")
    ?.addEventListener("click", handleSignup);
  // App navigation
  document.querySelectorAll(".nav-tab").forEach((tab) => {
    tab.addEventListener("click", (e) => {
      const tabName = e.currentTarget.dataset.tab;
      if (tabName) {
        switchTab(tabName);
      }
    });
  });
  document.getElementById("logoutBtn")?.addEventListener("click", handleLogout);
  // Feed filters
  document.querySelectorAll(".filter-btn").forEach((btn) => {
    btn.addEventListener("click", (e) => filterFeed(e.target.dataset.section));
  });
  // Upload form
  document
    .getElementById("uploadSubmit")
    ?.addEventListener("click", handleUpload);
  document
    .getElementById("uploadCaption")
    ?.addEventListener("input", updateCharCount);
  document
    .getElementById("uploadFile")
    ?.addEventListener("change", handleFileSelect);
  // File upload area drag and drop
  const fileUploadArea = document.getElementById("fileUploadArea");
  if (fileUploadArea) {
    fileUploadArea.addEventListener("dragover", (e) => {
      e.preventDefault();
      fileUploadArea.style.borderColor = "#667eea";
      fileUploadArea.style.background = "#f7fafc";
    });
    fileUploadArea.addEventListener("dragleave", (e) => {
      e.preventDefault();
      fileUploadArea.style.borderColor = "#cbd5e0";
      fileUploadArea.style.background = "";
    });
    fileUploadArea.addEventListener("drop", (e) => {
      e.preventDefault();
      fileUploadArea.style.borderColor = "#cbd5e0";
      fileUploadArea.style.background = "";
      const files = e.dataTransfer.files;
      if (files.length > 0) {
        document.getElementById("uploadFile").files = files;
        handleFileSelect({ target: { files: files } });
      }
    });
  }
  // Profile tabs
  document.querySelectorAll(".section-tab").forEach((tab) => {
    tab.addEventListener("click", (e) =>
      switchProfileSection(e.target.dataset.section)
    );
  });
  // Modal click outside to close
  document.getElementById("authModal")?.addEventListener("click", (e) => {
    if (e.target.id === "authModal") {
      hideAuthModal();
    }
  });
  console.log("[v0] Event listeners setup complete");
}

async function triggerNeighbourHelp() {
  try {
    showToast("Creating help request…", "info");
    const params = new URLSearchParams({ elderLat: "40.7128", elderLng: "-74.0060", elderId: "demo-elder" });
    const res = await fetch(`/api/upload?${params.toString()}`, { method: "POST" });
    const data = await res.json();
    if (res.ok) {
      showToast("Help request created!", "success");
    } else {
      showToast(data.message || "Failed to create request", "error");
    }
  } catch (e) {
    console.error(e);
    showToast("Network error", "error");
  }
}

async function handleChatbotAudioUpload(e) {
  const file = e.target.files[0];
  if (!file) return;
  if (!file.type.startsWith("audio/")) {
    showToast("Please select an audio file", "error");
    return;
  }
  try {
    showToast("Sending audio to chatbot…", "info");
    const form = new FormData();
    form.append("audio", file);
    const res = await fetch("/api/audio-chat", { method: "POST", body: form });
    const data = await res.json();
    if (res.ok) {
      showToast("Chatbot replied", "success");
      if (data.audioPath) {
        const audio = new Audio(data.audioPath);
        audio.play();
      }
    } else {
      showToast(data.error || "Chatbot failed", "error");
    }
  } catch (err) {
    console.error(err);
    showToast("Network error", "error");
  } finally {
    e.target.value = "";
  }
}

// Page management
function showHomepage() {
  console.log("[v0] Showing homepage");
  document.getElementById("homepage").classList.add("active");
  document.getElementById("mainApp").classList.remove("active");
}

function showMainApp() {
  console.log("[v0] Showing main app");
  document.getElementById("homepage").classList.remove("active");
  document.getElementById("mainApp").classList.add("active");
}

function scrollToFeatures() {
  document.querySelector(".features").scrollIntoView({ behavior: "smooth" });
}

// Auth modal management
function showAuthModal() {
  console.log("[v0] Showing auth modal");
  document.getElementById("authModal").classList.add("active");
  showLoginForm();
}

function hideAuthModal() {
  console.log("[v0] Hiding auth modal");
  document.getElementById("authModal").classList.remove("active");
}

function showLoginForm() {
  console.log("[v0] Showing login form");
  document.getElementById("authTitle").textContent = "Welcome Back";
  document.getElementById("loginForm").classList.add("active");
  document.getElementById("signupForm").classList.remove("active");
}

function showSignupForm() {
  console.log("[v0] Showing signup form");
  document.getElementById("authTitle").textContent = "Join Community Care";
  document.getElementById("loginForm").classList.remove("active");
  document.getElementById("signupForm").classList.add("active");
}

// Authentication
async function handleLogin() {
  console.log("[v0] Handling login");
  const email = document.getElementById("loginEmail").value.trim();
  const password = document.getElementById("loginPassword").value;
  console.log("[v0] Login attempt with email:", email);
  console.log("[v0] Password length:", password.length);
  if (!email || !password) {
    showToast("Please fill in all fields", "error");
    return;
  }
  try {
    console.log("[v0] Sending login request");
    const response = await fetch("/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    console.log("[v0] Response status:", response.status);
    console.log("[v0] Response ok:", response.ok);
    console.log(
      "[v0] Response headers:",
      Object.fromEntries(response.headers.entries())
    );
    const data = await response.json();
    console.log("[v0] Login response data:", data);
    const token = data.data ? data.data.token : data.token;
    const userId = data.data ? data.data.user_id : data.user_id;
    console.log("[v0] Extracted token:", token);
    console.log("[v0] Extracted user_id:", userId);
    console.log("[v0] Response has token:", !!token);
    if (response.ok && token) {
      localStorage.setItem("token", token);
      currentUser = { id: userId, email };
      console.log("[v0] Login successful, user:", currentUser);
      hideAuthModal();
      showMainApp();
      loadUserProfile();
      loadFeed();
      showToast("Welcome back! 🎉", "success");
    } else {
      console.error("[v0] Login failed - Response OK:", response.ok);
      console.error("[v0] Login failed - Has token:", !!token);
      console.error("[v0] Login failed - Full response:", data);
      console.error("[v0] Login failed - Response status:", response.status);
      // Show more specific error message
      let errorMessage = "Login failed. Please check your credentials.";
      if (data.message) {
        errorMessage = data.message;
      } else if (data.data && data.data.message) {
        errorMessage = data.data.message;
      } else if (!response.ok) {
        errorMessage = `Server error (${response.status}). Please try again.`;
      }
      showToast(errorMessage, "error");
    }
  } catch (error) {
    console.error("[v0] Login network error:", error);
    showToast("Network error. Please try again.", "error");
  }
}

async function handleSignup() {
  console.log("[v0] Handling signup");
  const name = document.getElementById("signupName").value.trim();
  const email = document.getElementById("signupEmail").value.trim();
  const password = document.getElementById("signupPassword").value;
  if (!name || !email || !password) {
    showToast("Please fill in all fields", "error");
    return;
  }
  if (password.length < 6) {
    showToast("Password must be at least 6 characters", "error");
    return;
  }
  try {
    console.log("[v0] Sending signup request");
    const response = await fetch("/signup", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, email, password }),
    });
    const data = await response.json();
    console.log("[v0] Signup response:", data);
    const token = data.data ? data.data.token : data.token;
    const userId = data.data ? data.data.user_id : data.user_id;
    if (response.ok && token) {
      localStorage.setItem("token", token);
      currentUser = { id: userId, name, email };
      console.log("[v0] Signup successful, user:", currentUser);
      hideAuthModal();
      showMainApp();
      loadUserProfile();
      loadFeed();
      showToast("Welcome to Community Care! 🎉", "success");
    } else {
      console.error("[v0] Signup failed:", data);
      let errorMessage = "Signup failed. Please try again.";
      if (data.message) {
        errorMessage = data.message;
      } else if (data.data && data.data.message) {
        errorMessage = data.data.message;
      }
      showToast(errorMessage, "error");
    }
  } catch (error) {
    console.error("[v0] Signup error:", error);
    showToast("Network error. Please try again.", "error");
  }
}

function handleLogout() {
  console.log("[v0] Handling logout");
  localStorage.removeItem("token");
  currentUser = null;
  showHomepage();
  showToast("Logged out successfully", "success");
}

// Tab management
function switchTab(tabName) {
  console.log("[v0] Switching to tab:", tabName);
  currentTab = tabName;
  // Update nav tabs
  document.querySelectorAll(".nav-tab").forEach((tab) => {
    tab.classList.remove("active");
  });
  const targetTab = document.querySelector(`[data-tab="${tabName}"]`);
  if (targetTab) {
    targetTab.classList.add("active");
  } else {
    console.error("[v0] Tab element not found for:", tabName);
  }
  // Update tab content
  document.querySelectorAll(".tab-content").forEach((content) => {
    content.classList.remove("active");
  });
  const targetContent = document.getElementById(`${tabName}Tab`);
  if (targetContent) {
    targetContent.classList.add("active");
  } else {
    console.error("[v0] Tab content not found for:", tabName);
  }
  // Load content based on tab
  if (tabName === "feed") {
    loadFeed();
  } else if (tabName === "profile") {
    loadUserProfile();
    loadUserPosts();
  }
}

// Feed management
function filterFeed(section) {
  console.log("[v0] Filtering feed by section:", section);
  currentSection = section;
  // Update filter buttons
  document.querySelectorAll(".filter-btn").forEach((btn) => {
    btn.classList.remove("active");
  });
  document.querySelector(`[data-section="${section}"]`).classList.add("active");
  loadFeed();
}

async function loadFeed() {
  console.log("[v0] Loading feed for section:", currentSection);
  const feedContainer = document.getElementById("feedContainer");
  feedContainer.innerHTML = '<div class="loading-spinner"></div>';
  try {
    const url =
      currentSection === "all" ? "/feed" : `/feed?section=${currentSection}`;
    console.log("[v0] Fetching feed from:", url);
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem("token")}`,
        "Content-Type": "application/json",
      },
    });
    console.log("[v0] Feed response status:", response.status);
    console.log("[v0] Feed response ok:", response.ok);
    const data = await response.json();
    console.log("[v0] Feed response data:", data);
    let posts = [];
    if (Array.isArray(data)) {
      posts = data;
    } else if (data.data && Array.isArray(data.data)) {
      posts = data.data;
    } else if (data.posts && Array.isArray(data.posts)) {
      posts = data.posts;
    } else {
      console.error("[v0] Unexpected feed response structure:", data);
      posts = [];
    }
    console.log("[v0] Feed loaded:", posts.length, "posts");
    console.log("[v0] Posts array:", posts);
    if (posts.length > 0) {
      feedContainer.innerHTML = posts
        .map((post) => createFeedItem(post))
        .join("");
    } else {
      feedContainer.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-heart" style="font-size: 4rem; color: #cbd5e0; margin-bottom: 20px;"></i>
                    <h3>No posts yet</h3>
                    <p>Be the first to share something with the community!</p>
                    <p style="font-size: 0.9em; color: #666; margin-top: 10px;">
                        ${
                          currentSection === "all"
                            ? "Try uploading some content!"
                            : `No ${currentSection} posts found.`
                        }
                    </p>
                </div>
            `;
    }
  } catch (error) {
    console.error("[v0] Error loading feed:", error);
    feedContainer.innerHTML = `
            <div class="error-state">
                <i class="fas fa-exclamation-triangle" style="font-size: 4rem; color: #e53e3e; margin-bottom: 20px;"></i>
                <h3>Failed to load feed</h3>
                <p>Please try again later.</p>
                <button onclick="loadFeed()" style="margin-top: 10px; padding: 8px 16px; background: #667eea; color: white; border: none; border-radius: 4px; cursor: pointer;">
                    Retry
                </button>
            </div>
        `;
  }
}

function createFeedItem(post) {
  const timeAgo = getTimeAgo(new Date(post.created_at));
  const userName = post.user_name || "Community Member";
  const userInitial = userName.charAt(0).toUpperCase();
  let mediaHtml = "";
  if (post.media_url) {
    if (post.media_type === "image") {
      mediaHtml = `
                <div class="feed-item-media">
                    <img src="${post.media_url}" alt="Post image" loading="lazy">
                </div>
            `;
    } else if (post.media_type === "video") {
      mediaHtml = `
                <div class="feed-item-media">
                    <video controls preload="metadata">
                        <source src="${post.media_url}" type="video/mp4">
                        Your browser does not support video playback.
                    </video>
                </div>
            `;
    } else if (post.media_type === "audio") {
      mediaHtml = `
                <div class="feed-item-media">
                    <div class="audio-player">
                        <i class="fas fa-music"></i>
                        <p>Audio Message</p>
                        <audio controls>
                            <source src="${post.media_url}" type="audio/mpeg">
                            Your browser does not support audio playback.
                        </audio>
                    </div>
                </div>
            `;
    }
  }

  // Display user's content (text/caption)
  const contentHtml = post.content
    ? `<div class="feed-item-text">${post.content}</div>`
    : "";

  // Display tags (optional)
  const tagsHtml =
    post.tags && post.tags.length > 0
      ? `
<div class="feed-item-tags">
${post.tags.map((tag) => `<span class="tag">${tag}</span>`).join("")}
</div>
`
      : "";
  return `
        <div class="feed-item">
            <div class="feed-item-header">
                <div class="user-avatar">${userInitial}</div>
                <div class="user-info">
                    <h4>${userName}</h4>
                    <span class="post-time">${timeAgo}</span>
                </div>
                <span class="section-badge ${post.section}">${
    post.section === "remedies" ? "🌿 Remedies" : "📖 Experience"
  }</span>
            </div>
            <div class="feed-item-content">
                ${mediaHtml}
                ${contentHtml}
                ${tagsHtml}
                <div class="feed-item-actions">
                    <button class="action-btn" onclick="toggleLike('${
                      post._id
                    }')">
                        <i class="fas fa-heart"></i>
                        <span>Like</span>
                    </button>
                    <button class="action-btn" onclick="sharePost('${
                      post._id
                    }')">
                        <i class="fas fa-share"></i>
                        <span>Share</span>
                    </button>
                    <button class="action-btn" onclick="savePost('${
                      post._id
                    }')">
                        <i class="fas fa-bookmark"></i>
                        <span>Save</span>
                    </button>
                </div>
            </div>
        </div>
    `;
}

// Upload management
function updateCharCount() {
  const textarea = document.getElementById("uploadCaption");
  const charCount = document.querySelector(".char-count");
  const current = textarea.value.length;
  const max = 500;
  charCount.textContent = `${current}/${max} characters`;
  if (current > max * 0.9) {
    charCount.style.color = "#e53e3e";
  } else {
    charCount.style.color = "#718096";
  }
}

function handleFileSelect(event) {
  console.log("[v0] File selected");
  const file = event.target.files[0];
  const preview = document.getElementById("filePreview");
  if (!file) {
    preview.classList.remove("active");
    return;
  }
  // Validate file size (50MB max)
  if (file.size > 50 * 1024 * 1024) {
    showToast("File size must be less than 50MB", "error");
    event.target.value = "";
    return;
  }
  // Validate file type
  const validTypes = ["image/", "video/", "audio/"];
  if (!validTypes.some((type) => file.type.startsWith(type))) {
    showToast("Please select an image, video, or audio file", "error");
    event.target.value = "";
    return;
  }
  // Show preview
  preview.classList.add("active");
  if (file.type.startsWith("image/")) {
    const reader = new FileReader();
    reader.onload = (e) => {
      preview.innerHTML = `
                <img src="${e.target.result}" alt="Preview">
                <div class="file-info">
                    <i class="fas fa-image"></i>
                    <span>${file.name} (${formatFileSize(file.size)})</span>
                </div>
            `;
    };
    reader.readAsDataURL(file);
  } else if (file.type.startsWith("video/")) {
    const reader = new FileReader();
    reader.onload = (e) => {
      preview.innerHTML = `
                <video controls>
                    <source src="${e.target.result}" type="${file.type}">
                </video>
                <div class="file-info">
                    <i class="fas fa-video"></i>
                    <span>${file.name} (${formatFileSize(file.size)})</span>
                </div>
            `;
    };
    reader.readAsDataURL(file);
  } else if (file.type.startsWith("audio/")) {
    preview.innerHTML = `
            <div class="audio-player">
                <i class="fas fa-music"></i>
                <p>Audio File Selected</p>
                <audio controls>
                    <source src="${URL.createObjectURL(file)}" type="${
      file.type
    }">
                </audio>
            </div>
            <div class="file-info">
                <i class="fas fa-music"></i>
                <span>${file.name} (${formatFileSize(file.size)})</span>
            </div>
        `;
  }
}

async function handleUpload() {
  console.log("[v0] Starting upload process");
  const caption = document.getElementById("uploadCaption").value.trim();
  const section = document.querySelector('input[name="section"]:checked').value;
  const file = document.getElementById("uploadFile").files[0];
  const token = localStorage.getItem("token");
  if (!token) {
    showToast("Please login first", "error");
    return;
  }
  // Validation based on section
  // Validation based on section
  if (section === "experience") {
    if (!file) {
      showToast(
        "Experience posts require media (image, video, or audio)",
        "error"
      );
      return;
    }
    if (file.type.startsWith("image/") && !caption) {
      showToast("Images must have a caption", "error");
      return;
    }
  } else if (section === "remedies") {
    // Fix this validation - remedies need EITHER text OR media, not both
    if (!file && !caption) {
      showToast("Remedies require either text or media", "error");
      return;
    }
    if (file && file.type.startsWith("image/") && !caption) {
      showToast("Images must have a caption", "error");
      return;
    }
  }

  // Add debug logging
  console.log("[v0] Upload validation - section:", section);
  console.log("[v0] Upload validation - file:", file);
  console.log("[v0] Upload validation - caption:", caption);
  console.log("[v0] Upload validation - caption length:", caption.length);

  // Validation based on section
  if (section === "experience") {
    if (!file) {
      console.log("[v0] Experience validation failed: no file");
      showToast(
        "Experience posts require media (image, video, or audio)",
        "error"
      );
      return;
    }
    if (file.type.startsWith("image/") && !caption) {
      console.log("[v0] Experience validation failed: image without caption");
      showToast("Images must have a caption", "error");
      return;
    }
  } else if (section === "remedies") {
    if (!file && !caption) {
      console.log("[v0] Remedies validation failed: no file and no caption");
      showToast("Remedies require either text or media", "error");
      return;
    }
    if (file && file.type.startsWith("image/") && !caption) {
      console.log("[v0] Remedies validation failed: image without caption");
      showToast("Images must have a caption", "error");
      return;
    }
  }

  console.log("[v0] Upload validation passed");
  // Validate media duration for video/audio (2 minutes max)
  if (
    file &&
    (file.type.startsWith("video/") || file.type.startsWith("audio/"))
  ) {
    const duration = await getMediaDuration(file);
    if (duration > 120) {
      // 2 minutes
      showToast("Video and audio files must be 2 minutes or less", "error");
      return;
    }
  }
  const uploadBtn = document.getElementById("uploadSubmit");
  const originalText = uploadBtn.innerHTML;
  uploadBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Uploading...';
  uploadBtn.disabled = true;
  try {
    let mediaUrl = "";
    let mediaType = "";
    // Upload to Cloudinary if file exists
    if (file) {
      console.log("[v0] Uploading file to Cloudinary");
      const formData = new FormData();
      formData.append("file", file);
      formData.append("upload_preset", UPLOAD_PRESET);
      formData.append("folder", `posts/${section}`);
      const cloudResponse = await fetch(
        `https://api.cloudinary.com/v1_1/${CLOUD_NAME}/auto/upload`,
        {
          method: "POST",
          body: formData,
        }
      );
      const cloudData = await cloudResponse.json();
      console.log("[v0] Cloudinary response:", cloudData);
      if (!cloudData.secure_url) {
        throw new Error("Failed to upload media");
      }
      mediaUrl = cloudData.secure_url;
      mediaType = file.type.startsWith("video/")
        ? "video"
        : file.type.startsWith("audio/")
        ? "audio"
        : "image";
      console.log("[v0] Media uploaded successfully:", mediaUrl);
    }
    // Send to backend
    console.log("[v0] Sending post to backend");
    const postData = {
      token,
      media_url: mediaUrl,
      media_type: mediaType,
      content: caption,
      tags: [],
      section,
    };
    const response = await fetch("/posts", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        //Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(postData),
    });
    const result = await response.json();
    console.log("[v0] Backend response:", result);
    if (response.ok) {
      console.log("[v0] Post created successfully");
      showToast("Post shared successfully! 🎉", "success");
      // Reset form
      document.getElementById("uploadCaption").value = "";
      document.getElementById("uploadFile").value = "";
      document.getElementById("filePreview").classList.remove("active");
      updateCharCount();
      // Switch to feed and reload
      switchTab("feed");
    } else {
      console.error("[v0] Post creation failed:", result);
      showToast(result.message || "Failed to create post", "error");
    }
  } catch (error) {
    console.error("[v0] Upload error:", error);
    showToast("Upload failed. Please try again.", "error");
  } finally {
    uploadBtn.innerHTML = originalText;
    uploadBtn.disabled = false;
  }
}

// Profile management
async function loadUserProfile() {
  console.log("[v0] Loading user profile");
  const token = localStorage.getItem("token");
  if (!token) return;
  try {
    // For now, use stored user data
    // In a real app, you'd fetch from /profile endpoint
    if (currentUser) {
      document.getElementById("profileName").textContent =
        currentUser.name || "Community Member";
      document.getElementById("profileEmail").textContent =
        currentUser.email || "";
      // Set avatar initial
      const initial = (currentUser.name || currentUser.email || "U")
        .charAt(0)
        .toUpperCase();
      document.getElementById("profileAvatar").innerHTML = initial;
    }
    // Load stats (mock data for now)
    document.getElementById("remediesCount").textContent = "0";
    document.getElementById("experienceCount").textContent = "0";
    document.getElementById("helpCount").textContent = "0";
  } catch (error) {
    console.error("[v0] Error loading profile:", error);
  }
}

function switchProfileSection(section) {
  console.log("[v0] Switching profile section to:", section);
  // Update tabs
  document.querySelectorAll(".section-tab").forEach((tab) => {
    tab.classList.remove("active");
  });
  document.querySelector(`[data-section="${section}"]`).classList.add("active");
  loadUserPosts(section);
}

async function loadUserPosts(section = "remedies") {
  console.log("[v0] Loading user posts for section:", section);
  const postsContainer = document.getElementById("profilePosts");
  postsContainer.innerHTML = '<div class="loading-spinner"></div>';
  try {
    const token = localStorage.getItem("token");
    if (!token) {
      throw new Error("No authentication token");
    }

    const response = await fetch(`/user/posts?section=${section}`, {
      method: "POST", // Change to POST to send token in body
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ token }),
    });
    console.log("[v0] User posts response status:", response.status);
    if (response.ok) {
      const data = await response.json();
      console.log("[v0] User posts data:", data);
      // Handle different response structures
      let posts = [];
      if (Array.isArray(data)) {
        posts = data;
      } else if (data.data && Array.isArray(data.data)) {
        posts = data.data;
      } else if (data.posts && Array.isArray(data.posts)) {
        posts = data.posts;
      }
      if (posts.length > 0) {
        postsContainer.innerHTML = posts
          .map((post) => createFeedItem(post))
          .join("");
        // Update counts
        document.getElementById(`${section}Count`).textContent = posts.length;
      } else {
        postsContainer.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-${
                          section === "remedies" ? "leaf" : "book"
                        }" style="font-size: 3rem; color: #cbd5e0; margin-bottom: 15px;"></i>
                        <h4>No ${section} shared yet</h4>
                        <p>Start sharing your ${
                          section === "remedies"
                            ? "traditional remedies"
                            : "life experiences"
                        } with the community!</p>
                    </div>
                `;
      }
    } else {
      throw new Error(`HTTP ${response.status}`);
    }
  } catch (error) {
    console.error("[v0] Error loading user posts:", error);
    postsContainer.innerHTML = `
            <div class="error-state">
                <i class="fas fa-exclamation-triangle" style="font-size: 3rem; color: #e53e3e; margin-bottom: 15px;"></i>
                <h4>Failed to load posts</h4>
                <p>Please try again later.</p>
                <button onclick="loadUserPosts('${section}')" style="margin-top: 10px; padding: 8px 16px; background: #667eea; color: white; border: none; border-radius: 4px; cursor: pointer;">
                    Retry
                </button>
            </div>
        `;
  }
}

// Utility functions
function showToast(message, type = "success") {
  console.log("[v0] Showing toast:", message, type);
  const container = document.getElementById("toastContainer");
  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.innerHTML = `
        <div style="display: flex; align-items: center; gap: 10px;">
            <i class="fas fa-${
              type === "success"
                ? "check-circle"
                : type === "error"
                ? "exclamation-circle"
                : "info-circle"
            }"></i>
            <span>${message}</span>
        </div>
    `;
  container.appendChild(toast);
  setTimeout(() => {
    toast.remove();
  }, 5000);
}

function getTimeAgo(date) {
  const now = new Date();
  const diffInSeconds = Math.floor((now - date) / 1000);
  if (diffInSeconds < 60) return "Just now";
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
  if (diffInSeconds < 604800)
    return `${Math.floor(diffInSeconds / 86400)}d ago`;
  return date.toLocaleDateString();
}

function formatFileSize(bytes) {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return (
    Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i]
  );
}

function getMediaDuration(file) {
  return new Promise((resolve) => {
    const media = file.type.startsWith("video/")
      ? document.createElement("video")
      : document.createElement("audio");
    media.preload = "metadata";
    media.onloadedmetadata = () => {
      resolve(media.duration);
    };
    media.onerror = () => {
      resolve(0); // If we can't get duration, allow upload
    };
    media.src = URL.createObjectURL(file);
  });
}

// Placeholder functions for feed actions
function toggleLike(postId) {
  console.log("[v0] Toggle like for post:", postId);
  showToast("Like feature coming soon!", "info");
}

function sharePost(postId) {
  console.log("[v0] Share post:", postId);
  showToast("Share feature coming soon!", "info");
}

function savePost(postId) {
  console.log("[v0] Save post:", postId);
  showToast("Save feature coming soon!", "info");
}

// Helper function for API calls
async function postJSON(url, body, token) {
  const headers = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = `Bearer ${token}`;
  const response = await fetch(url, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  });
  return response.json();
}

console.log("[v0] App.js loaded successfully");
