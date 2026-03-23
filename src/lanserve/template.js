
  const dropZone = document.getElementById('drop-zone');
  const fileInput = document.getElementById('file-input');
  const fileName  = document.getElementById('file-name');

  fileInput.addEventListener('change', () => {
    fileName.textContent = fileInput.files[0]?.name || '';
  });
  dropZone.addEventListener('dragover', e => { e.preventDefault(); dropZone.classList.add('drag-over'); });
  dropZone.addEventListener('dragleave', () => dropZone.classList.remove('drag-over'));
  dropZone.addEventListener('drop', e => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
    if (e.dataTransfer.files.length) {
      fileInput.files = e.dataTransfer.files;
      fileName.textContent = fileInput.files[0].name;
    }
  });

  function toast(msg, isError = false) {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.className = isError ? 'error show' : 'show';
    setTimeout(() => t.className = '', 3000);
  }

  function doUpload() {
    const file = fileInput.files[0];
    if (!file) { toast('Pick a file first', true); return; }
    const folder = document.getElementById('folder-select').value;
    const fd = new FormData();
    fd.append('file', file);
    fd.append('target_folder', folder);

    const xhr = new XMLHttpRequest();
    xhr.open('POST', window.location.pathname);

    const progressWrap = document.getElementById('progress-wrap');
    const progressBar  = document.getElementById('progress-bar');
    const status       = document.getElementById('upload-status');

    progressWrap.style.display = 'block';
    xhr.upload.onprogress = e => {
      if (e.lengthComputable) {
        const pct = Math.round(e.loaded / e.total * 100);
        progressBar.style.width = pct + '%';
        status.textContent = `Uploading… ${pct}%`;
      }
    };
    xhr.onload = () => {
      progressWrap.style.display = 'none';
      progressBar.style.width = '0%';
      if (xhr.status === 200 || xhr.status === 204) {
        toast('✓ Uploaded: ' + file.name);
        status.textContent = '';
        fileInput.value = '';
        fileName.textContent = '';
        setTimeout(() => location.reload(), 800);
      } else {
        toast('Upload failed (' + xhr.status + ')', true);
        status.textContent = '';
      }
    };
    xhr.onerror = () => { toast('Network error', true); progressWrap.style.display = 'none'; };
    xhr.send(fd);
  }

  function deleteFile(encodedPath, name) {
    if (!confirm(`Delete "${name}"?`)) return;
    fetch(encodedPath, { method: 'DELETE' })
      .then(r => {
        if (r.status === 204) {
          toast('🗑 Deleted: ' + name);
          setTimeout(() => location.reload(), 600);
        } else {
          toast('Delete failed (' + r.status + ')', true);
        }
      })
      .catch(() => toast('Network error', true));
  }
