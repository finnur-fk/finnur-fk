// Constants
const STATUS_HIDE_DELAY_MS = 5000;
const ANIMATION_DURATION_MS = 300;

document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('uploadForm');
    const uploadArea = document.getElementById('uploadArea');
    const fileInput = document.getElementById('csvfile');
    const uploadButton = document.getElementById('uploadButton');
    const uploadStatus = document.getElementById('uploadStatus');

    // Handle drag and drop
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        uploadArea.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    ['dragenter', 'dragover'].forEach(eventName => {
        uploadArea.addEventListener(eventName, () => {
            uploadArea.classList.add('dragover');
        });
    });

    ['dragleave', 'drop'].forEach(eventName => {
        uploadArea.addEventListener(eventName, () => {
            uploadArea.classList.remove('dragover');
        });
    });

    uploadArea.addEventListener('drop', function(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        
        if (files.length > 0) {
            fileInput.files = files;
            updateFileName(files[0].name);
        }
    });

    // Update file name display when file is selected
    fileInput.addEventListener('change', function(e) {
        if (e.target.files.length > 0) {
            updateFileName(e.target.files[0].name);
        }
    });

    function updateFileName(name) {
        const p = uploadArea.querySelector('p');
        p.textContent = `Selected: ${name}`;
        p.style.color = '#10b981';
        p.style.fontWeight = '600';
    }

    // Handle form submission
    uploadForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        if (!fileInput.files || fileInput.files.length === 0) {
            showStatus('Please select a CSV file to upload', 'error');
            return;
        }

        const file = fileInput.files[0];
        
        // Validate file type
        if (!file.name.toLowerCase().endsWith('.csv')) {
            showStatus('Please upload a valid CSV file', 'error');
            return;
        }

        // Prepare form data
        const formData = new FormData();
        formData.append('csvfile', file);

        // Disable button during upload
        uploadButton.disabled = true;
        uploadButton.textContent = 'Uploading...';

        try {
            const response = await fetch('/upload', {
                method: 'POST',
                body: formData
            });

            const result = await response.json();

            if (response.ok && result.success) {
                showStatus(`✅ Success! File "${result.filename}" has been securely uploaded to the vault.`, 'success');
                
                // Reset form
                uploadForm.reset();
                const p = uploadArea.querySelector('p');
                p.textContent = 'or click to browse';
                p.style.color = '';
                p.style.fontWeight = '';
                
                // Update transaction count (demo purposes)
                updateTransactionCount();
            } else {
                showStatus(`❌ Upload failed: ${result.message || 'Unknown error'}`, 'error');
            }
        } catch (error) {
            console.error('Upload error:', error);
            showStatus(`❌ Upload failed: ${error.message || 'Network error'}`, 'error');
        } finally {
            uploadButton.disabled = false;
            uploadButton.textContent = 'Upload to Vault';
        }
    });

    function showStatus(message, type) {
        uploadStatus.textContent = message;
        uploadStatus.className = `upload-status show ${type}`;
        
        // Auto-hide after delay
        setTimeout(() => {
            uploadStatus.classList.remove('show');
        }, STATUS_HIDE_DELAY_MS);
    }

    function updateTransactionCount() {
        const statValue = document.querySelector('.stat-card .stat-value');
        if (statValue) {
            const currentValue = parseInt(statValue.textContent) || 0;
            const newValue = currentValue > 0 ? currentValue + 1 : 604;
            statValue.textContent = newValue + '+';
            
            // Add animation
            statValue.style.transform = 'scale(1.2)';
            statValue.style.transition = `transform ${ANIMATION_DURATION_MS}ms ease`;
            setTimeout(() => {
                statValue.style.transform = 'scale(1)';
            }, ANIMATION_DURATION_MS);
        }
    }
});
