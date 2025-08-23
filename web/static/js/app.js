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
                // Always update preview in split-pane mode
                this.updatePreview();
                // Trigger cursor sync after preview update
                setTimeout(() => {
                    if (typeof this.syncToCursor === 'function') {
                        this.syncToCursor();
                    }
                }, 50);
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
        
        // Delete confirmation modal
        const confirmDeleteBtn = document.getElementById('confirm-delete');
        if (confirmDeleteBtn) {
            confirmDeleteBtn.addEventListener('click', () => {
                this.closeModals();
                this.confirmDeleteNote();
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
        const savedTheme = localStorage.getItem('ml-notes-theme') || 'dark';
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
    
    deleteCurrentNote() {
        if (!this.currentNoteId) return;
        
        // Show custom delete confirmation modal
        this.showDeleteModal();
    }

    async confirmDeleteNote() {
        if (!this.currentNoteId) return;
        
        try {
            const response = await fetch(`/api/v1/notes/${this.currentNoteId}`, {
                method: 'DELETE'
            });
            
            const data = await response.json();
            
            if (data.success) {
                this.showNotification('Note deleted successfully');
                window.location.href = '/';
            } else {
                this.showNotification(data.error || 'Failed to delete note', 'error');
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
        // Set up split pane functionality
        this.setupSplitPane();
        // Initialize live preview
        this.updatePreview();
    }
    
    setupSplitPane() {
        const container = document.getElementById('split-pane-container');
        const editorPane = document.getElementById('editor-pane');
        const resizeHandle = document.getElementById('split-resize-handle');
        const textarea = document.getElementById('note-content');
        const preview = document.getElementById('note-preview');
        
        if (!container || !editorPane || !resizeHandle || !textarea || !preview) return;
        
        let isResizing = false;
        let startX = 0;
        let startWidth = 0;
        let isScrollSyncing = false; // Prevent infinite scroll loops
        
        // Set initial state - preview focused (editor hidden)
        container.classList.add('focus-preview');
        
        // Handle editor focus - expand to 50%
        textarea.addEventListener('focus', () => {
            if (!isResizing) {
                container.classList.remove('focus-preview');
                container.classList.add('focus-editor');
            }
        });
        
        // Handle editor blur - but only shrink if clicking outside the editor area
        textarea.addEventListener('blur', (e) => {
            // Small delay to check if focus moved to resize handle or stays in editor area
            setTimeout(() => {
                const activeElement = document.activeElement;
                const isInEditorArea = activeElement === textarea || 
                                     activeElement === resizeHandle ||
                                     editorPane.contains(activeElement);
                
                if (!isInEditorArea && !isResizing) {
                    container.classList.remove('focus-editor');
                    container.classList.add('focus-preview');
                }
            }, 100);
        });
        
        // Synchronized scrolling functionality
        const syncScroll = (source, target) => {
            if (isScrollSyncing) return;
            isScrollSyncing = true;
            
            const sourceScrollPercent = source.scrollTop / (source.scrollHeight - source.clientHeight);
            const targetScrollTop = sourceScrollPercent * (target.scrollHeight - target.clientHeight);
            
            target.scrollTop = Math.max(0, targetScrollTop);
            
            // Reset the flag after a brief delay
            setTimeout(() => {
                isScrollSyncing = false;
            }, 10);
        };
        
        // Auto-scroll to cursor position when content changes
        this.syncToCursor = () => {
            if (isScrollSyncing) return;
            
            const cursorPosition = textarea.selectionStart;
            const textBeforeCursor = textarea.value.substring(0, cursorPosition);
            const linesBeforeCursor = textBeforeCursor.split('\n').length;
            
            // Get line height approximation
            const style = window.getComputedStyle(textarea);
            const lineHeight = parseInt(style.lineHeight) || parseInt(style.fontSize) * 1.2;
            
            // Calculate cursor position in pixels
            const cursorTopPosition = (linesBeforeCursor - 1) * lineHeight;
            
            // Check if content extends beyond editor's visible area
            const editorContentHeight = textarea.scrollHeight;
            const editorVisibleHeight = textarea.clientHeight;
            const contentExtendsEditor = editorContentHeight > editorVisibleHeight;
            
            // Check if cursor is near the bottom of the visible area
            const visibleBottom = textarea.scrollTop + textarea.clientHeight;
            const cursorNearBottom = cursorTopPosition > (visibleBottom - lineHeight * 2);
            
            if (contentExtendsEditor && cursorNearBottom) {
                // Scroll preview to the bottom when content extends editor
                preview.scrollTop = preview.scrollHeight - preview.clientHeight;
                
                // Also scroll editor if cursor goes beyond visible area
                if (cursorTopPosition > textarea.scrollTop + textarea.clientHeight - lineHeight) {
                    textarea.scrollTop = cursorTopPosition - textarea.clientHeight + lineHeight * 2;
                }
            }
        };
        
        // Sync preview scroll when editor scrolls
        textarea.addEventListener('scroll', () => {
            syncScroll(textarea, preview);
        });
        
        // Sync editor scroll when preview scrolls
        preview.addEventListener('scroll', () => {
            syncScroll(preview, textarea);
        });
        
        // Track cursor movement and key events for auto-scroll
        textarea.addEventListener('keydown', (e) => {
            // Trigger sync on Enter key (new line) and navigation keys
            if (e.key === 'Enter' || e.key === 'ArrowDown' || e.key === 'ArrowUp' || 
                e.key === 'PageDown' || e.key === 'PageUp' || e.key === 'End' || e.key === 'Home') {
                setTimeout(() => {
                    this.syncToCursor();
                }, 10);
            }
        });
        
        // Track cursor position changes via mouse clicks
        textarea.addEventListener('click', () => {
            setTimeout(() => {
                this.syncToCursor();
            }, 10);
        });
        
        // Manual resize functionality
        resizeHandle.addEventListener('mousedown', (e) => {
            e.preventDefault();
            isResizing = true;
            startX = e.clientX;
            startWidth = editorPane.offsetWidth;
            
            // Remove focus classes during manual resize
            container.classList.remove('focus-editor', 'focus-preview');
            
            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);
            
            // Add resizing class for visual feedback
            document.body.style.cursor = 'ew-resize';
            container.classList.add('resizing');
        });
        
        function handleMouseMove(e) {
            if (!isResizing) return;
            
            const containerWidth = container.offsetWidth - resizeHandle.offsetWidth;
            const deltaX = e.clientX - startX;
            const newWidth = startWidth + deltaX;
            
            // Constrain width between 15% and 85%
            const minWidth = containerWidth * 0.15;
            const maxWidth = containerWidth * 0.85;
            const constrainedWidth = Math.max(minWidth, Math.min(maxWidth, newWidth));
            
            const percentage = (constrainedWidth / containerWidth) * 100;
            editorPane.style.width = `${percentage}%`;
        }
        
        function handleMouseUp() {
            isResizing = false;
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            
            // Reset cursor and remove resizing class
            document.body.style.cursor = '';
            container.classList.remove('resizing');
            
            // Don't auto-adjust after manual resize - let user's choice persist
        }
        
        // Double-click to toggle between 50/50 and preview-focused
        resizeHandle.addEventListener('dblclick', () => {
            const currentWidth = editorPane.offsetWidth;
            const containerWidth = container.offsetWidth - resizeHandle.offsetWidth;
            const currentPercentage = (currentWidth / containerWidth) * 100;
            
            // If close to 50%, go to preview mode (0%), otherwise go to 50%
            if (Math.abs(currentPercentage - 50) < 10) {
                container.classList.add('focus-preview');
                container.classList.remove('focus-editor');
                editorPane.style.width = ''; // Reset to CSS-controlled width
            } else {
                container.classList.add('focus-editor');
                container.classList.remove('focus-preview');
                editorPane.style.width = ''; // Reset to CSS-controlled width
            }
        });
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
            const analysisText = result.analysis || result.summary || 'Analysis completed successfully.';
            
            // Render markdown if marked.js is available
            if (window.marked) {
                let html = marked.parse(analysisText);
                
                // Sanitize HTML if DOMPurify is available
                if (window.DOMPurify) {
                    html = DOMPurify.sanitize(html);
                }
                
                contentDiv.innerHTML = html;
            } else {
                // Fallback to plain text if marked.js is not available
                contentDiv.textContent = analysisText;
            }
        }
        
        resultDiv.style.display = 'block';
        
        // If analysis was written back, reload the page to show updates
        if (result.written_back || result.new_note_id) {
            setTimeout(() => {
                window.location.reload();
            }, 2000);
        }
    }
    
    showDeleteModal() {
        const modal = document.getElementById('delete-modal');
        const overlay = document.getElementById('modal-overlay');
        const titleElement = document.getElementById('delete-preview-title');
        
        if (modal && overlay) {
            // Get the current note title
            const noteTitle = document.getElementById('note-title')?.value || 'Untitled Note';
            if (titleElement) {
                titleElement.textContent = noteTitle;
            }
            
            overlay.style.display = 'block';
            modal.style.display = 'block';
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
            fontWeight: '500',
            zIndex: '9999',
            maxWidth: '300px',
            boxShadow: '0 4px 12px rgba(var(--shadow-color), 0.15)',
            transform: 'translateX(100%)',
            transition: 'transform 0.3s ease',
            border: '1px solid var(--border-color)'
        });
        
        // Set colors based on type using CSS custom properties
        switch(type) {
            case 'success':
                notification.style.background = 'var(--success-color)';
                notification.style.color = 'white';
                break;
            case 'error':
                notification.style.background = 'var(--danger-color)';
                notification.style.color = 'white';
                break;
            case 'warning':
                notification.style.background = '#f59e0b';
                notification.style.color = 'white';
                break;
            case 'info':
            default:
                notification.style.background = 'var(--modal-bg)';
                notification.style.color = 'var(--text-primary)';
                notification.style.boxShadow = '0 4px 12px rgba(var(--shadow-color), 0.15), 0 0 0 1px var(--border-color)';
                break;
        }
        
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