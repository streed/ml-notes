#!/usr/bin/env python3
"""
OpenWebUI Integration Example for ML Notes

This script demonstrates how to integrate ml-notes with OpenWebUI using custom functions.
Copy these functions into your OpenWebUI Functions section to enable note-taking capabilities.

Prerequisites:
1. Start the ml-notes HTTP server: ./ml-notes serve --host 0.0.0.0 --port 8080
2. Configure the base URL below to match your setup
3. Add these functions to OpenWebUI

Usage in OpenWebUI chat:
- "Create a note about machine learning basics"
- "Search my notes for python examples" 
- "What notes do I have about AI?"
- "Auto-tag my recent notes"
"""

import json
import requests
from typing import Optional, List, Dict, Any

# Configuration - Update this to match your ml-notes server
ML_NOTES_BASE_URL = "http://localhost:8080/api/v1"

class MLNotesAPI:
    def __init__(self, base_url: str = ML_NOTES_BASE_URL):
        self.base_url = base_url.rstrip('/')
    
    def _make_request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        """Make HTTP request to ml-notes API"""
        url = f"{self.base_url}{endpoint}"
        try:
            response = requests.request(method, url, timeout=30, **kwargs)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            return {"success": False, "error": str(e)}
    
    def health_check(self) -> Dict[str, Any]:
        """Check if ml-notes server is healthy"""
        return self._make_request("GET", "/health")
    
    def search_notes(self, query: str, limit: int = 5, use_vector: bool = True) -> Dict[str, Any]:
        """Search notes using text or vector search"""
        data = {
            "query": query,
            "limit": limit,
            "use_vector": use_vector
        }
        return self._make_request("POST", "/notes/search", json=data)
    
    def create_note(self, title: str, content: str, tags: str = "", auto_tag: bool = True) -> Dict[str, Any]:
        """Create a new note with optional auto-tagging"""
        data = {
            "title": title,
            "content": content,
            "tags": tags,
            "auto_tag": auto_tag
        }
        return self._make_request("POST", "/notes", json=data)
    
    def get_note(self, note_id: int) -> Dict[str, Any]:
        """Get a specific note by ID"""
        return self._make_request("GET", f"/notes/{note_id}")
    
    def list_notes(self, limit: int = 10, offset: int = 0) -> Dict[str, Any]:
        """List recent notes"""
        params = {"limit": limit, "offset": offset}
        return self._make_request("GET", "/notes", params=params)
    
    def suggest_tags(self, note_id: int) -> Dict[str, Any]:
        """Get AI-suggested tags for a note"""
        return self._make_request("POST", f"/auto-tag/suggest/{note_id}")
    
    def auto_tag_notes(self, note_ids: List[int] = None, all_notes: bool = False, 
                      recent: int = 0, apply: bool = False) -> Dict[str, Any]:
        """Apply auto-tagging to notes"""
        data = {
            "apply": apply,
            "overwrite": False
        }
        
        if all_notes:
            data["all"] = True
        elif recent > 0:
            data["recent"] = recent
        elif note_ids:
            data["note_ids"] = note_ids
        else:
            return {"success": False, "error": "Must specify notes to tag"}
        
        return self._make_request("POST", "/auto-tag/apply", json=data)
    
    def get_stats(self) -> Dict[str, Any]:
        """Get database statistics"""
        return self._make_request("GET", "/stats")

# Initialize API client
ml_notes = MLNotesAPI()

def test_connection():
    """Test connection to ml-notes server"""
    result = ml_notes.health_check()
    if result.get("success"):
        print("✅ Connected to ml-notes server successfully!")
        stats = ml_notes.get_stats()
        if stats.get("success"):
            data = stats["data"]
            print(f"📊 Database has {data['total_notes']} notes and {data['total_tags']} tags")
            print(f"🔍 Vector search: {'enabled' if data['vector_search'] else 'disabled'}")
            print(f"🤖 Auto-tagging: {'enabled' if data['auto_tagging'] else 'disabled'}")
    else:
        print(f"❌ Failed to connect: {result.get('error', 'Unknown error')}")
        print(f"Make sure ml-notes server is running on {ML_NOTES_BASE_URL}")

# OpenWebUI Function Templates
# Copy these into your OpenWebUI Functions section

OPENWEBUI_FUNCTIONS = """
# Function 1: Create Note
def create_note_function(title: str, content: str, tags: str = "", user_message: str = "", **kwargs) -> str:
    \"\"\"
    Create a new note with AI-powered auto-tagging.
    
    Args:
        title: The title of the note
        content: The main content of the note  
        tags: Optional comma-separated tags
        user_message: The original user message for context
    \"\"\"
    import requests
    import json
    
    api_url = "http://localhost:8080/api/v1/notes"
    
    data = {
        "title": title,
        "content": content,
        "tags": tags,
        "auto_tag": True
    }
    
    try:
        response = requests.post(api_url, json=data, timeout=30)
        response.raise_for_status()
        result = response.json()
        
        if result.get("success"):
            note = result["data"]
            tags_str = ", ".join(note.get("tags", []))
            return f"✅ Created note '{note['title']}' (ID: {note['id']})\\nTags: {tags_str}"
        else:
            return f"❌ Failed to create note: {result.get('error', 'Unknown error')}"
    
    except Exception as e:
        return f"❌ Error creating note: {str(e)}"

# Function 2: Search Notes  
def search_notes_function(query: str, limit: int = 5, **kwargs) -> str:
    \"\"\"
    Search your notes using AI-powered vector search or text search.
    
    Args:
        query: What to search for in your notes
        limit: Maximum number of results to return
    \"\"\"
    import requests
    import json
    
    api_url = "http://localhost:8080/api/v1/notes/search"
    
    data = {
        "query": query,
        "limit": limit,
        "use_vector": True
    }
    
    try:
        response = requests.post(api_url, json=data, timeout=30)
        response.raise_for_status()
        result = response.json()
        
        if result.get("success"):
            notes = result["data"]
            if not notes:
                return f"No notes found matching '{query}'"
            
            response_text = f"Found {len(notes)} notes matching '{query}':\\n\\n"
            
            for i, note in enumerate(notes, 1):
                tags_str = ", ".join(note.get("tags", []))
                content_preview = note["content"][:150]
                if len(note["content"]) > 150:
                    content_preview += "..."
                
                response_text += f"{i}. **{note['title']}** (ID: {note['id']})\\n"
                response_text += f"   Tags: {tags_str}\\n"
                response_text += f"   {content_preview}\\n\\n"
            
            return response_text
        else:
            return f"❌ Search failed: {result.get('error', 'Unknown error')}"
    
    except Exception as e:
        return f"❌ Error searching notes: {str(e)}"

# Function 3: Auto-tag Recent Notes
def auto_tag_recent_function(count: int = 5, apply: bool = False, **kwargs) -> str:
    \"\"\"
    Use AI to automatically suggest or apply tags to your recent notes.
    
    Args:
        count: Number of recent notes to process
        apply: Whether to actually apply the tags (default: False for preview)
    \"\"\"
    import requests
    import json
    
    api_url = "http://localhost:8080/api/v1/auto-tag/apply"
    
    data = {
        "recent": count,
        "apply": apply,
        "overwrite": False
    }
    
    try:
        response = requests.post(api_url, json=data, timeout=60)
        response.raise_for_status()
        result = response.json()
        
        if result.get("success"):
            data = result["data"]
            action = "Applied" if apply else "Suggested"
            response_text = f"🤖 {action} AI tags for {data['processed_count']} notes:\\n\\n"
            
            for i, note_result in enumerate(data["results"], 1):
                if note_result.get("success"):
                    suggested = ", ".join(note_result.get("suggested_tags", []))
                    final = ", ".join(note_result.get("final_tags", []))
                    
                    response_text += f"{i}. **{note_result['note_title']}**\\n"
                    response_text += f"   Suggested: {suggested}\\n"
                    if apply:
                        response_text += f"   Applied: {final}\\n"
                    response_text += "\\n"
                else:
                    response_text += f"{i}. **{note_result['note_title']}**: {note_result.get('error', 'Failed')}\\n\\n"
            
            if not apply:
                response_text += "💡 To actually apply these tags, use apply=True"
            
            return response_text
        else:
            return f"❌ Auto-tagging failed: {result.get('error', 'Unknown error')}"
    
    except Exception as e:
        return f"❌ Error auto-tagging notes: {str(e)}"

# Function 4: List Recent Notes
def list_notes_function(limit: int = 10, **kwargs) -> str:
    \"\"\"
    List your most recent notes.
    
    Args:
        limit: Number of notes to show
    \"\"\"
    import requests
    
    api_url = f"http://localhost:8080/api/v1/notes?limit={limit}"
    
    try:
        response = requests.get(api_url, timeout=30)
        response.raise_for_status()
        result = response.json()
        
        if result.get("success"):
            notes = result["data"]
            if not notes:
                return "No notes found"
            
            response_text = f"📝 Your {len(notes)} most recent notes:\\n\\n"
            
            for i, note in enumerate(notes, 1):
                tags_str = ", ".join(note.get("tags", []))
                content_preview = note["content"][:100]
                if len(note["content"]) > 100:
                    content_preview += "..."
                
                response_text += f"{i}. **{note['title']}** (ID: {note['id']})\\n"
                response_text += f"   Tags: {tags_str}\\n"
                response_text += f"   Created: {note['created_at'][:10]}\\n"
                response_text += f"   {content_preview}\\n\\n"
            
            return response_text
        else:
            return f"❌ Failed to list notes: {result.get('error', 'Unknown error')}"
    
    except Exception as e:
        return f"❌ Error listing notes: {str(e)}"
"""

def main():
    """Test the ml-notes integration"""
    print("🚀 ML Notes + OpenWebUI Integration Test")
    print("=" * 50)
    
    # Test connection
    test_connection()
    print()
    
    # Show example usage
    print("💡 Example usage in OpenWebUI:")
    print("- 'Create a note about machine learning fundamentals'")
    print("- 'Search my notes for python examples'")
    print("- 'Show me my recent notes'") 
    print("- 'Auto-tag my 5 most recent notes'")
    print()
    
    print("📋 OpenWebUI Functions to copy:")
    print("Copy the functions from OPENWEBUI_FUNCTIONS variable above")
    print("into your OpenWebUI Functions section.")

if __name__ == "__main__":
    main()