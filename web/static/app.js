const dropZone = document.getElementById("dropZone");

["dragenter", "dragover", "dragleave", "drop"].forEach((eventName) => {
  dropZone.addEventListener(eventName, (e) => e.preventDefault());
  dropZone.addEventListener(eventName, (e) => e.stopPropagation());
});

["dragenter", "dragover"].forEach((eventName) => {
  dropZone.addEventListener(eventName, () =>
    dropZone.classList.add("dragover"),
  );
});
["dragleave", "drop"].forEach((eventName) => {
  dropZone.addEventListener(eventName, () =>
    dropZone.classList.remove("dragover"),
  );
});

dropZone.addEventListener("drop", (e) => {
  const dt = e.dataTransfer;
  uploadFiles(dt.files);
});

function filterFiles() {
  const q = document.getElementById("searchInput").value.toLowerCase();
  const rows = document.querySelectorAll(".file-row");
  rows.forEach((row) => {
    const name = row.getAttribute("data-name");
    row.style.display = name.includes(q) ? "" : "none";
  });
}

function uploadFiles(files) {
  if (!files.length) return;

  const formData = new FormData();
  const targetFolder = document.getElementById("targetFolder").value;
  formData.append("target_folder", targetFolder);

  for (let i = 0; i < files.length; i++) {
    formData.append("file", files[i]);
  }

  const xhr = new XMLHttpRequest();
  const progressContainer = document.getElementById("progressContainer");
  const progressBar = document.getElementById("progressBar");
  const progressText = document.getElementById("progressText");

  progressContainer.classList.remove("hidden");

  xhr.upload.addEventListener("progress", (e) => {
    if (e.lengthComputable) {
      const percent = Math.round((e.loaded / e.total) * 100);
      progressBar.style.width = percent + "%";
      progressText.innerText = `Uploading: ${percent}%`;
    }
  });

  xhr.onload = () => {
    if (xhr.status === 204 || xhr.status === 200) {
      progressText.innerText = "Upload complete! Refreshing...";
      setTimeout(() => window.location.reload(), 500);
    } else if (xhr.status === 401) {
      const code = prompt("Access Code Required:");
      if (code) {
        xhr.open("POST", "/");
        xhr.setRequestHeader("X-Auth-Code", code);
        xhr.send(formData);
      } else {
        progressContainer.classList.add("hidden");
      }
    } else {
      alert("Upload failed: " + xhr.responseText);
      progressContainer.classList.add("hidden");
    }
  };

  xhr.open("POST", "/");
  xhr.send(formData);
}

function deleteFile(path, name) {
  if (!confirm(`Are you sure you want to delete "${name}"?`)) return;

  const executeDelete = (code = null) => {
    const xhr = new XMLHttpRequest();
    xhr.open("DELETE", path);
    if (code) xhr.setRequestHeader("X-Auth-Code", code);

    xhr.onload = () => {
      if (xhr.status === 204) {
        window.location.reload();
      } else if (xhr.status === 401) {
        const newCode = prompt("Access Code Required:");
        if (newCode) executeDelete(newCode);
      } else {
        alert("Delete failed");
      }
    };
    xhr.send();
  };

  executeDelete();
}
