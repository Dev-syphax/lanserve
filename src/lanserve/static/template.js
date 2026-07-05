const dropZone = document.getElementById("drop-zone");
const fileInput = document.getElementById("file-input");
const fileName = document.getElementById("file-name");

function describeFiles(files) {
  if (!files.length) return "";
  if (files.length === 1) return files[0].name;
  return `${files.length} files selected`;
}

fileInput.addEventListener("change", () => {
  fileName.textContent = describeFiles(fileInput.files);
});
dropZone.addEventListener("dragover", (e) => {
  e.preventDefault();
  dropZone.classList.add("drag-over");
});
dropZone.addEventListener("dragleave", () =>
  dropZone.classList.remove("drag-over"),
);
dropZone.addEventListener("drop", (e) => {
  e.preventDefault();
  dropZone.classList.remove("drag-over");
  if (e.dataTransfer.files.length) {
    fileInput.files = e.dataTransfer.files;
    fileName.textContent = describeFiles(e.dataTransfer.files);
  }
});

function toast(msg, isError = false) {
  const t = document.getElementById("toast");
  t.textContent = msg;
  t.className = isError ? "error show" : "show";
  setTimeout(() => (t.className = ""), 3000);
}

function refreshFileList() {
  fetch(window.location.href)
    .then((r) => r.text())
    .then((html) => {
      const doc = new DOMParser().parseFromString(html, "text/html");
      const newList = doc.querySelector(".file-list");
      const newFolder = doc.querySelector("#folder-select");
      if (newList)
        document.querySelector(".file-list").innerHTML = newList.innerHTML;
      if (newFolder)
        document.getElementById("folder-select").innerHTML =
          newFolder.innerHTML;
    })
    .catch(() => location.reload());
}

function doUpload() {
  const files = fileInput.files;
  if (!files.length) {
    toast("Pick a file first", true);
    return;
  }
  const folder = document.getElementById("folder-select").value;
  attemptUpload(files, folder, null);
}

function attemptUpload(files, folder, code) {
  const fd = new FormData();
  for (const file of files) fd.append("file", file);
  fd.append("target_folder", folder);

  const xhr = new XMLHttpRequest();
  xhr.open("POST", window.location.pathname);
  if (code) xhr.setRequestHeader("X-Auth-Code", code);

  const progressWrap = document.getElementById("progress-wrap");
  const progressBar = document.getElementById("progress-bar");
  const status = document.getElementById("upload-status");

  progressWrap.style.display = "block";
  xhr.upload.onprogress = (e) => {
    if (e.lengthComputable) {
      const pct = Math.round((e.loaded / e.total) * 100);
      progressBar.style.width = pct + "%";
      status.textContent = `Uploading… ${pct}%`;
    }
  };
  xhr.onload = () => {
    progressWrap.style.display = "none";
    progressBar.style.width = "0%";
    if (xhr.status === 200 || xhr.status === 204) {
      const label =
        files.length === 1 ? files[0].name : `${files.length} files`;
      toast("✓ Uploaded: " + label);
      status.textContent = "";
      fileInput.value = "";
      fileName.textContent = "";
      setTimeout(refreshFileList, 800);
    } else if (xhr.status === 401) {
      progressWrap.style.display = "none";
      if (code !== null) {
        toast("Wrong access code", true);
        status.textContent = "";
        return;
      }
      const entered = window.prompt("Enter access code:");
      if (entered === null) {
        toast("Access code required", true);
        status.textContent = "";
        return;
      }
      attemptUpload(files, folder, entered);
    } else {
      toast("Upload failed (" + xhr.status + ")", true);
      status.textContent = "";
    }
  };
  xhr.onerror = () => {
    toast("Network error", true);
    progressWrap.style.display = "none";
  };
  xhr.send(fd);
}

function deleteFile(encodedPath, name) {
  if (!confirm(`Delete "${name}"?`)) return;
  attemptDelete(encodedPath, name, null);
}

function attemptDelete(encodedPath, name, code) {
  fetch(encodedPath, {
    method: "DELETE",
    headers: code ? { "X-Auth-Code": code } : {},
  })
    .then((r) => {
      if (r.status === 204) {
        toast("🗑 Deleted: " + name);
        setTimeout(refreshFileList, 600);
      } else if (r.status === 401) {
        if (code !== null) {
          toast("Wrong access code", true);
          return;
        }
        const entered = window.prompt("Enter access code:");
        if (entered === null) {
          toast("Access code required", true);
          return;
        }
        attemptDelete(encodedPath, name, entered);
      } else {
        toast("Delete failed (" + r.status + ")", true);
      }
    })
    .catch(() => toast("Network error", true));
}
