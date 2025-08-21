// ML Notes Web App JavaScript

class MLNotesApp {
    constructor() {
        this.currentNoteId = null;
        this.isNewNote = false;
        this.isPreviewMode = false;
        this.debounceTimer = null;
        this.unsavedChanges = false;
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.setupTheme();
        this.loadCurrentNote();
        this.setupAutoSave();
        this.setupSearch();
        this.setupMarkdownPreview();
    }
    
    setupEventListeners() {
        // Theme toggle
        document.getElementById('theme-toggle').addEventListener('click', () => {
            this.toggleTheme();
        });
        
        // New note button
        document.getElementById('new-note-btn').addEventListener('click', () => {
            window.location.href = '/new';
        });
        
        // Create first note button (welcome screen)
        const createFirstNoteBtn = document.getElementById('create-first-note');
        if (createFirstNoteBtn) {
            createFirstNoteBtn.addEventListener('click', () => {
                window.location.href = '/new';
            });
        }
        
        // Note list items
        document.querySelectorAll('.note-item').forEach(item => {
            item.addEventListener('click', () => {
                const noteId = parseInt(item.dataset.noteId);
                this.loadNote(noteId);
            });
        });
        
        // Editor controls
        const saveBtn = document.getElementById('save-note');
        if (saveBtn) {
            saveBtn.addEventListener('click', () => {
                this.saveCurrentNote();
            });
        }
        
        const deleteBtn = document.getElementById('delete-note');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', () => {
                this.deleteCurrentNote();
            });
        }
        
        const previewBtn = document.getElementById('toggle-preview');
        if (previewBtn) {
            previewBtn.addEventListener('click', () => {
                this.togglePreview();
            });
        }
        
        const autoTagBtn = document.getElementById('auto-tag-btn');
        if (autoTagBtn) {
            autoTagBtn.addEventListener('click', () => {
                this.autoTagNote();
            });
        }
        
        const analyzeBtn = document.getElementById('analyze-btn');
        if (analyzeBtn) {
            analyzeBtn.addEventListener('click', () => {
                this.showAnalysisModal();
            });
        }
        
        // Content change tracking
        const titleInput = document.getElementById('note-title');
        const contentTextarea = document.getElementById('note-content');
        const tagsInput = document.getElementById('note-tags');
        
        if (titleInput) {
            titleInput.addEventListener('input', () => {
                this.markUnsaved();
                this.updateDocumentTitle();
            });
        }
        
        if (contentTextarea) {
            contentTextarea.addEventListener('input', () => {
                this.markUnsaved();
                if (this.isPreviewMode) {
                    this.updatePreview();
                }
            });
        }
        
        if (tagsInput) {
            tagsInput.addEventListener('input', () => {
                this.markUnsaved();
            });
        }
        
        // Tag removal
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('tag-remove')) {
                const tag = e.target.closest('.tag');
                const tagValue = tag.dataset.tag;
                this.removeTag(tagValue);
            }
        });
        
        // Search
        const searchInput = document.getElementById('search-input');
        const searchBtn = document.getElementById('search-btn');
        
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.performSearch(e.target.value);
            });
            
            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch(e.target.value);
                }
            });
        }
        
        if (searchBtn) {
            searchBtn.addEventListener('click', () => {
                const query = searchInput.value;
                this.performSearch(query);
            });
        }
        
        // Tag filter
        const tagFilter = document.getElementById('tag-filter');
        if (tagFilter) {
            tagFilter.addEventListener('change', (e) => {
                this.filterByTag(e.target.value);
            });
        }
        
        // Modal controls
        this.setupModalControls();
        
        // Keyboard shortcuts
        this.setupKeyboardShortcuts();
        
        // Before unload warning
        window.addEventListener('beforeunload', (e) => {
            if (this.unsavedChanges) {
                e.preventDefault();
                e.returnValue = '';
            }
        });
    }
    
    setupModalControls() {
        // Close modals
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-overlay') || 
                e.target.classList.contains('modal-close')) {
                this.closeModals();
            }
        });
        
        // Analysis modal
        const analysisWriteBack = document.getElementById('analysis-write-back');
        const analysisWriteNew = document.getElementById('analysis-write-new');
        const analysisTitleInput = document.getElementById('analysis-title-input');
        
        if (analysisWriteBack && analysisWriteNew) {
            analysisWriteBack.addEventListener('change', () => {
                if (analysisWriteBack.checked) {
                    analysisWriteNew.checked = false;
                    analysisTitleInput.style.display = 'none';
                }
            });
            
            analysisWriteNew.addEventListener('change', () => {
                if (analysisWriteNew.checked) {
                    analysisWriteBack.checked = false;
                    analysisTitleInput.style.display = 'block';
                } else {
                    analysisTitleInput.style.display = 'none';
                }
            });
        }
        
        const runAnalysisBtn = document.getElementById('run-analysis');
        if (runAnalysisBtn) {
            runAnalysisBtn.addEventListener('click', () => {
                this.runAnalysis();
            });
        }
    }
    
    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + S = Save
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.saveCurrentNote();
            }
            
            // Ctrl/Cmd + N = New note
            if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
                e.preventDefault();
                window.location.href = '/new';
            }
            
            // Ctrl/Cmd + P = Toggle preview
            if ((e.ctrlKey || e.metaKey) && e.key === 'p') {
                e.preventDefault();
                this.togglePreview();
            }
            
            // Ctrl/Cmd + / = Toggle theme
            if ((e.ctrlKey || e.metaKey) && e.key === '/') {
                e.preventDefault();
                this.toggleTheme();
            }
            
            // Escape = Close modals
            if (e.key === 'Escape') {
                this.closeModals();
            }
        });
    }
    
    setupTheme() {
        const savedTheme = localStorage.getItem('ml-notes-theme') || 'light';
        this.setTheme(savedTheme);
    }
    
    setTheme(theme) {
        document.body.dataset.theme = theme;
        localStorage.setItem('ml-notes-theme', theme);
        
        const themeIcon = document.getElementById('theme-icon');
        if (themeIcon) {
            themeIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
        }
    }
    
    toggleTheme() {
        const currentTheme = document.body.dataset.theme;
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        this.setTheme(newTheme);
    }
    
    loadCurrentNote() {
        const currentNoteId = document.getElementById('current-note-id');
        const isNewNoteInput = document.getElementById('is-new-note');
        
        if (currentNoteId && currentNoteId.value) {
            this.currentNoteId = parseInt(currentNoteId.value) || null;
            this.updateDocumentTitle();
        }
        
        if (isNewNoteInput) {
            this.isNewNote = isNewNoteInput.value === 'true';
        }
    }
    
    updateDocumentTitle() {
        const titleInput = document.getElementById('note-title');
        if (titleInput && titleInput.value) {
            document.title = `${titleInput.value} - ML Notes`;
        } else {
            document.title = 'ML Notes';
        }
    }
    
    async loadNote(noteId) {
        try {
            const response = await fetch(`/api/v1/notes/${noteId}`);
            const data = await response.json();
            
            if (data.success) {
                window.location.href = `/note/${noteId}`;
            } else {
                this.showNotification('Failed to load note', 'error');
            }
        } catch (error) {
            console.error('Error loading note:', error);
            this.showNotification('Failed to load note', 'error');
        }
    }
    
    
    async saveCurrentNote() {
        const titleInput = document.getElementById('note-title');
        const contentTextarea = document.getElementById('note-content');
        const tagsInput = document.getElementById('note-tags');
        
        if (!titleInput || !contentTextarea) return;
        
        if (!titleInput.value.trim()) {
            this.showNotification('Please enter a title for your note', 'warning');
            titleInput.focus();
            return;
        }
        
        const noteData = {
            title: titleInput.value,
            content: contentTextarea.value,
            tags: tagsInput ? tagsInput.value : '',
            auto_tag: false
        };
        
        try {
            let response;
            
            if (this.isNewNote || !this.currentNoteId) {
                // Create new note
                response = await fetch('/api/v1/notes', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(noteData)
                });
            } else {
                // Update existing note
                response = await fetch(`/api/v1/notes/${this.currentNoteId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(noteData)
                });
            }
            
            const data = await response.json();
            
            if (data.success) {
                this.markSaved();
                if (this.isNewNote || !this.currentNoteId) {
                    // Redirect to the new note
                    this.showNotification('Note created successfully', 'success');
                    setTimeout(() => {
                        window.location.href = `/note/${data.data.id}`;
                    }, 1000);
                } else {
                    this.showNotification('Note saved successfully', 'success');
                    this.updateNoteTags(data.data.tags);
                }
            } else {
                this.showNotification('Failed to save note', 'error');
            }
        } catch (error) {
            console.error('Error saving note:', error);
            this.showNotification('Failed to save note', 'error');
        }
    }
    
    async deleteCurrentNote() {
        if (!this.currentNoteId) return;
        
        if (!confirm('Are you sure you want to delete this note?')) return;
        
        try {
            const response = await fetch(`/api/v1/notes/${this.currentNoteId}`, {
                method: 'DELETE'
            });
            
            const data = await response.json();
            
            if (data.success) {
                window.location.href = '/';
            } else {
                this.showNotification('Failed to delete note', 'error');
            }
        } catch (error) {
            console.error('Error deleting note:', error);
            this.showNotification('Failed to delete note', 'error');
        }
    }
    
    setupAutoSave() {
        setInterval(() => {
            if (this.unsavedChanges && this.currentNoteId && !this.isNewNote) {
                this.saveCurrentNote();
            }
        }, 30000); // Auto-save every 30 seconds
    }
    
    markUnsaved() {
        this.unsavedChanges = true;
        const saveBtn = document.getElementById('save-note');
        if (saveBtn) {
            saveBtn.textContent = 'Save*';
            saveBtn.classList.add('unsaved');
        }
    }
    
    markSaved() {
        this.unsavedChanges = false;
        const saveBtn = document.getElementById('save-note');
        if (saveBtn) {
            saveBtn.textContent = 'Save';
            saveBtn.classList.remove('unsaved');
        }
    }
    
    setupSearch() {
        // Search functionality is handled in event listeners
    }
    
    async performSearch(query) {
        if (!query.trim()) {
            this.clearSearch();
            return;
        }
        
        try {
            const searchData = {
                query: query,
                limit: 20,
                use_vector: true // Use vector search if available
            };
            
            const response = await fetch('/api/v1/notes/search', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(searchData)
            });
            
            const data = await response.json();
            
            if (data.success) {
                this.displaySearchResults(data.data);
            } else {
                this.showNotification('Search failed', 'error');
            }
        } catch (error) {
            console.error('Error searching:', error);
            this.showNotification('Search failed', 'error');
        }
    }
    
    displaySearchResults(notes) {
        const notesList = document.getElementById('notes-list');
        if (!notesList) return;
        
        // Clear current notes
        notesList.innerHTML = '';
        
        if (notes.length === 0) {
            notesList.innerHTML = '<div class=\"empty-state\"><p>No notes found.</p></div>';
            return;
        }
        
        notes.forEach(note => {
            const noteItem = this.createNoteElement(note);
            notesList.appendChild(noteItem);
        });
    }
    
    clearSearch() {
        // Reload the page to show all notes
        window.location.reload();
    }
    
    filterByTag(tag) {
        const noteItems = document.querySelectorAll('.note-item');
        
        noteItems.forEach(item => {
            const noteTags = item.dataset.tags.toLowerCase();
            if (!tag || noteTags.includes(tag.toLowerCase())) {
                item.style.display = 'block';
            } else {
                item.style.display = 'none';
            }
        });
    }
    
    createNoteElement(note) {
        const noteItem = document.createElement('div');
        noteItem.className = 'note-item';
        noteItem.dataset.noteId = note.id;
        noteItem.dataset.tags = note.tags ? note.tags.join(' ') : '';
        
        const preview = note.content.length > 100 ? note.content.substring(0, 100) + '...' : note.content;
        const createdDate = new Date(note.created_at).toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'short', 
            day: 'numeric' 
        });
        
        let tagsHtml = '';
        if (note.tags && note.tags.length > 0) {
            tagsHtml = '<div class=\"note-tags\">' + 
                note.tags.map(tag => `<span class=\"tag\">${tag}</span>`).join('') + 
                '</div>';
        }
        
        noteItem.innerHTML = `
            <div class=\"note-title\">${note.title}</div>
            <div class=\"note-preview\">${preview}</div>
            <div class=\"note-meta\">
                <span class=\"note-date\">${createdDate}</span>
                ${tagsHtml}
            </div>
        `;
        
        noteItem.addEventListener('click', () => {
            this.loadNote(note.id);
        });
        
        return noteItem;
    }
    
    setupMarkdownPreview() {
        // Preview functionality is set up in togglePreview
    }
    
    togglePreview() {
        const textarea = document.getElementById('note-content');
        const preview = document.getElementById('note-preview');
        const toggleBtn = document.getElementById('toggle-preview');
        const previewIcon = document.getElementById('preview-icon');
        const previewText = document.getElementById('preview-text');
        
        if (!textarea || !preview || !toggleBtn) return;
        
        this.isPreviewMode = !this.isPreviewMode;
        
        if (this.isPreviewMode) {
            textarea.style.display = 'none';
            preview.style.display = 'block';
            previewIcon.textContent = '✏️';
            previewText.textContent = 'Edit';
            this.updatePreview();
        } else {
            textarea.style.display = 'block';
            preview.style.display = 'none';
            previewIcon.textContent = '👁️';
            previewText.textContent = 'Preview';
        }
    }
    
    updatePreview() {
        const textarea = document.getElementById('note-content');
        const preview = document.getElementById('note-preview');
        
        if (!textarea || !preview || !window.marked) return;
        
        const markdown = textarea.value;
        let html = marked.parse(markdown);
        
        // Sanitize HTML if DOMPurify is available
        if (window.DOMPurify) {
            html = DOMPurify.sanitize(html);
        }
        
        preview.innerHTML = html;
    }
    
    async autoTagNote() {
        if (!this.currentNoteId) return;
        
        const autoTagBtn = document.getElementById('auto-tag-btn');
        if (autoTagBtn) {
            autoTagBtn.disabled = true;
            autoTagBtn.textContent = '🤖 Generating...';
        }
        
        try {
            const response = await fetch(`/api/v1/auto-tag/suggest/${this.currentNoteId}`, {
                method: 'POST'
            });
            
            const data = await response.json();
            
            if (data.success) {
                const suggestedTags = data.data.suggested_tags;
                if (suggestedTags && suggestedTags.length > 0) {
                    this.addSuggestedTags(suggestedTags);
                    this.showNotification(`Added ${suggestedTags.length} auto-generated tags`, 'success');
                } else {
                    this.showNotification('No tags suggested', 'info');
                }
            } else {
                this.showNotification('Auto-tagging failed', 'error');
            }
        } catch (error) {
            console.error('Error auto-tagging:', error);
            this.showNotification('Auto-tagging failed', 'error');
        } finally {
            if (autoTagBtn) {
                autoTagBtn.disabled = false;
                autoTagBtn.textContent = '🏷️ Auto-tag';
            }
        }
    }
    
    addSuggestedTags(suggestedTags) {
        const tagsInput = document.getElementById('note-tags');
        if (!tagsInput) return;
        
        const currentTags = tagsInput.value.split(',').map(tag => tag.trim()).filter(tag => tag);
        const newTags = [...new Set([...currentTags, ...suggestedTags])];
        
        tagsInput.value = newTags.join(', ');
        this.updateCurrentTags(newTags);
        this.markUnsaved();
    }
    
    updateCurrentTags(tags) {
        const currentTagsContainer = document.getElementById('current-tags');
        if (!currentTagsContainer) return;
        
        currentTagsContainer.innerHTML = '';
        
        tags.forEach(tag => {
            if (tag.trim()) {
                const tagElement = document.createElement('span');
                tagElement.className = 'tag removable';
                tagElement.dataset.tag = tag.trim();
                tagElement.innerHTML = `${tag.trim()} <span class=\"tag-remove\">×</span>`;
                currentTagsContainer.appendChild(tagElement);
            }
        });
    }
    
    updateNoteTags(tags) {
        const tagsInput = document.getElementById('note-tags');
        if (tagsInput && tags) {
            tagsInput.value = tags.join(', ');
            this.updateCurrentTags(tags);
        }
    }
    
    removeTag(tagToRemove) {
        const tagsInput = document.getElementById('note-tags');
        if (!tagsInput) return;
        
        const currentTags = tagsInput.value.split(',').map(tag => tag.trim()).filter(tag => tag && tag !== tagToRemove);
        tagsInput.value = currentTags.join(', ');
        this.updateCurrentTags(currentTags);
        this.markUnsaved();
    }
    
    showAnalysisModal() {
        const modal = document.getElementById('analysis-modal');
        const overlay = document.getElementById('modal-overlay');
        
        if (modal && overlay) {
            overlay.style.display = 'block';
            modal.style.display = 'block';
            
            // Reset modal state
            document.getElementById('analysis-write-back').checked = false;
            document.getElementById('analysis-write-new').checked = false;
            document.getElementById('analysis-title-input').style.display = 'none';
            document.getElementById('analysis-prompt').value = '';
            document.getElementById('analysis-result').style.display = 'none';
        }
    }
    
    async runAnalysis() {
        if (!this.currentNoteId) return;
        
        const writeBack = document.getElementById('analysis-write-back').checked;
        const writeNew = document.getElementById('analysis-write-new').checked;
        const customTitle = document.getElementById('analysis-title').value;
        const prompt = document.getElementById('analysis-prompt').value;
        
        const runBtn = document.getElementById('run-analysis');
        runBtn.disabled = true;
        runBtn.textContent = 'Analyzing...';
        
        try {
            // Build query parameters
            const params = new URLSearchParams();
            if (writeBack) params.append('write-back', 'true');
            if (writeNew) params.append('write-new', 'true');
            if (customTitle) params.append('write-title', customTitle);
            if (prompt) params.append('prompt', prompt);
            
            // Use the CLI analyze endpoint (we'll need to add this to the API)
            const response = await fetch(`/api/v1/analyze/${this.currentNoteId}?${params}`, {
                method: 'POST'
            });
            
            const data = await response.json();
            
            if (data.success) {
                this.showAnalysisResult(data.data);
            } else {
                this.showNotification('Analysis failed: ' + (data.error || 'Unknown error'), 'error');
            }
        } catch (error) {
            console.error('Error running analysis:', error);
            this.showNotification('Analysis failed', 'error');
        } finally {
            runBtn.disabled = false;
            runBtn.textContent = 'Analyze';
        }
    }
    
    showAnalysisResult(result) {
        const resultDiv = document.getElementById('analysis-result');
        const contentDiv = resultDiv.querySelector('.analysis-content');
        
        if (contentDiv) {
            contentDiv.textContent = result.analysis || result.summary || 'Analysis completed successfully.';
        }
        
        resultDiv.style.display = 'block';
        
        // If analysis was written back, reload the page to show updates
        if (result.written_back || result.new_note_id) {
            setTimeout(() => {
                window.location.reload();
            }, 2000);
        }
    }
    
    closeModals() {
        const overlay = document.getElementById('modal-overlay');
        const modals = document.querySelectorAll('.modal');
        
        if (overlay) {
            overlay.style.display = 'none';
        }
        
        modals.forEach(modal => {
            modal.style.display = 'none';
        });
    }
    
    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        
        // Style the notification
        Object.assign(notification.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '12px 20px',
            borderRadius: '8px',
            color: 'white',
            fontWeight: '500',
            zIndex: '9999',
            maxWidth: '300px',
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
            transform: 'translateX(100%)',
            transition: 'transform 0.3s ease'
        });
        
        // Set background color based on type
        const colors = {
            success: '#10b981',
            error: '#ef4444',
            warning: '#f59e0b',
            info: '#3b82f6'
        };
        notification.style.backgroundColor = colors[type] || colors.info;
        
        // Add to page
        document.body.appendChild(notification);
        
        // Animate in
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
        }, 100);
        
        // Remove after delay
        setTimeout(() => {
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new MLNotesApp();
});