import React, { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import {
  Mic,
  Search,
  FolderKanban,
  ClosedCaption,
  Settings,
  Eye,
  ChevronDown,
  ChevronRight,
  HelpCircle,
} from "lucide-react";

interface HelpDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const HelpDialog: React.FC<HelpDialogProps> = ({
  open,
  onOpenChange,
}) => {
  const [openSections, setOpenSections] = useState<string[]>([]);

  const toggleSection = (sectionId: string) => {
    setOpenSections((prev) =>
      prev.includes(sectionId)
        ? prev.filter((id) => id !== sectionId)
        : [...prev, sectionId]
    );
  };

  const helpSections = [
    {
      id: "getting-started",
      title: "🚀 Getting Started",
      icon: <FolderKanban className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Quick Setup</h4>
            <ol className="list-decimal list-inside space-y-2 text-sm">
              <li>Create your first board from the "Manage Boards" section</li>
              <li>
                Add columns to organize your workflow (e.g., "To Do", "In
                Progress", "Done")
              </li>
              <li>
                Start creating tasks either manually or through voice commands
              </li>
              <li>Configure your transcription settings for voice features</li>
            </ol>
          </div>
          <div>
            <h4 className="font-medium mb-2">First Steps</h4>
            <p className="text-sm text-muted-foreground">
              Seisami is a voice-first productivity tool that combines Kanban
              boards with AI-powered voice transcription. Start by setting up
              your OpenAI API key in Settings to enable voice features.
            </p>
          </div>
        </div>
      ),
    },
    {
      id: "voice-commands",
      title: "🎤 Voice Commands & Transcription",
      icon: <Mic className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">How Voice Commands Work</h4>
            <p className="text-sm text-muted-foreground mb-3">
              Speak naturally to create tasks, organize your workflow, and
              manage your projects. The AI understands context and can perform
              multiple actions from a single command.
            </p>
          </div>
          <div>
            <h4 className="font-medium mb-2">Transcription Features</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>Real-time voice-to-text conversion</li>
              <li>AI-powered intent recognition</li>
              <li>Automatic task creation and organization</li>
              <li>View all transcriptions in the Transcripts tab</li>
            </ul>
          </div>
        </div>
      ),
    },
    {
      id: "kanban-boards",
      title: "📋 Managing Kanban Boards",
      icon: <FolderKanban className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Board Management</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                <strong>Create Boards:</strong> Use "Manage Boards" to create
                new project boards
              </li>
              <li>
                <strong>Switch Boards:</strong> Use the board selector in the
                sidebar
              </li>
              <li>
                <strong>Organize Columns:</strong> Add, edit, and delete columns
                to match your workflow
              </li>
              <li>
                <strong>Card Management:</strong> Create, edit, move, and delete
                task cards
              </li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Working with Cards</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                <strong>Click to View:</strong> Click any card to see full
                details
              </li>
              <li>
                <strong>Edit Descriptions:</strong> Add detailed descriptions to
                tasks
              </li>
              <li>
                <strong>Drag & Drop:</strong> Move cards between columns by
                dragging
              </li>
              <li>
                <strong>Quick Actions:</strong> Use the three-dot menu for quick
                actions
              </li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Column Features</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>Rename columns by clicking the edit icon</li>
              <li>Delete empty columns with the trash icon</li>
              <li>Columns wrap to new lines on smaller screens</li>
              <li>Color-coded for easy visual organization</li>
            </ul>
          </div>
        </div>
      ),
    },
    {
      id: "search-navigation",
      title: "🔍 Search & Navigation",
      icon: <Search className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Search Features</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                <strong>Deep Search:</strong> Search across card titles,
                descriptions, and column names
              </li>
              <li>
                <strong>Real-time Results:</strong> See results as you type
              </li>
              <li>
                <strong>Visual Highlighting:</strong> Matching text is
                highlighted in yellow
              </li>
              <li>
                <strong>Smart Navigation:</strong> Jump between search results
                easily
              </li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Keyboard Shortcuts</h4>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="font-mono">
                  Ctrl/Cmd + Enter
                </Badge>
                <span className="text-sm">Next search result</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="font-mono">
                  Shift + Enter
                </Badge>
                <span className="text-sm">Previous search result</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="font-mono">
                  Escape
                </Badge>
                <span className="text-sm">Clear search</span>
              </div>
            </div>
          </div>

          <div>
            <h4 className="font-medium mb-2">Search Tips</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>Use partial words to find matches</li>
              <li>Search is case-insensitive</li>
              <li>Use the counter to see total results</li>
              <li>Cards auto-scroll into view when selected</li>
            </ul>
          </div>
        </div>
      ),
    },
    {
      id: "transcripts",
      title: "📝 Viewing Transcriptions",
      icon: <ClosedCaption className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Transcripts Overview</h4>
            <p className="text-sm text-muted-foreground mb-3">
              The Transcripts tab shows all your voice recordings and their
              AI-processed results.
            </p>
          </div>

          <div>
            <h4 className="font-medium mb-2">Understanding Transcripts</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                <strong>Original Text:</strong> What you said (truncated to 200
                characters in list view)
              </li>
              <li>
                <strong>Intent Recognition:</strong> AI's understanding of what
                you wanted to do
              </li>
              <li>
                <strong>Processing Status:</strong> Green checkmark shows AI has
                processed the request
              </li>
              <li>
                <strong>Timestamps:</strong> When each recording was made
              </li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Transcript Actions</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>Click any transcript to view full details</li>
              <li>See complete AI response and actions taken</li>
              <li>Review what tasks were created or modified</li>
              <li>Use as reference for future voice commands</li>
            </ul>
          </div>
        </div>
      ),
    },
    {
      id: "settings",
      title: "⚙️ Settings & Configuration",
      icon: <Settings className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Transcription Options</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                <strong>OpenAI Whisper:</strong> Cloud-based transcription
                (recommended)
              </li>
              <li>
                <strong>Local Whisper:</strong> Offline transcription (requires
                setup). Configure binary and model paths for local setup
              </li>
              <li>Use your own OpenAI Api key</li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Getting Your OpenAI API Key</h4>
            <ol className="list-decimal list-inside space-y-1 text-sm">
              <li>
                Visit{" "}
                <a
                  href="https://platform.openai.com/api-keys"
                  className="text-blue-600 hover:underline"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  OpenAI API Keys
                </a>
              </li>
              <li>Create a new API key</li>
              <li>Copy the key and paste it in Seisami Settings</li>
              <li>Save settings to enable voice features</li>
            </ol>
          </div>
        </div>
      ),
    },
    {
      id: "tips-tricks",
      title: "💡 Tips & Best Practices",
      icon: <Eye className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Voice Command Tips</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>Speak clearly and at a normal pace</li>
              <li>Be specific about column names</li>
              <li>Use natural language - the AI understands context</li>
              <li>You can create multiple tasks in one command</li>
              <li>Mention deadlines and priorities when relevant</li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Workflow Organization</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                Use consistent column naming (e.g., "To Do", "In Progress",
                "Done")
              </li>
              <li>Keep task titles concise but descriptive</li>
              <li>Use descriptions for detailed requirements</li>
              <li>Regular cleanup of completed tasks</li>
            </ul>
          </div>

          <div>
            <h4 className="font-medium mb-2">Productivity Hacks</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>
                Use voice commands for quick task creation during meetings
              </li>
              <li>Search to quickly find specific tasks</li>
              <li>Review transcripts to track your productivity patterns</li>
              <li>Create templates using consistent voice commands</li>
            </ul>
          </div>
        </div>
      ),
    },
    {
      id: "troubleshooting",
      title: "🔧 Troubleshooting",
      icon: <Settings className="h-4 w-4" />,
      content: (
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Common Issues</h4>
            <div className="space-y-3">
              <div>
                <p className="text-sm font-medium">
                  Voice transcription not working
                </p>
                <ul className="list-disc list-inside text-xs text-muted-foreground ml-2">
                  <li>Check if OpenAI API key is set in Settings</li>
                  <li>Ensure microphone permissions are granted</li>
                  <li>Verify internet connection for cloud transcription</li>
                </ul>
              </div>
              <div>
                <p className="text-sm font-medium">
                  Tasks not being created from voice
                </p>
                <ul className="list-disc list-inside text-xs text-muted-foreground ml-2">
                  <li>Make sure you have at least one board created</li>
                  <li>Be specific about column names in your command</li>
                  <li>Check transcripts tab to see what the AI understood</li>
                </ul>
              </div>
              <div>
                <p className="text-sm font-medium">
                  Search not finding results
                </p>
                <ul className="list-disc list-inside text-xs text-muted-foreground ml-2">
                  <li>Try partial keywords instead of full phrases</li>
                  <li>Check spelling of search terms</li>
                  <li>
                    Search includes card titles, descriptions, and column names
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div>
            <h4 className="font-medium mb-2">Performance Tips</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>Close other applications while using voice features</li>
              <li>Use a good quality microphone for better transcription</li>
              <li>Keep boards organized to improve AI task placement</li>
            </ul>
          </div>
        </div>
      ),
    },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-semibold flex items-center gap-2">
            <HelpCircle className="h-5 w-5" />
            Seisami Help & Guide
          </DialogTitle>
        </DialogHeader>

        <div className="mt-4">
          <p className="text-muted-foreground text-sm mb-6">
            Welcome to Seisami! This guide will help you master voice-powered
            productivity and kanban board management.
          </p>

          <div className="space-y-2">
            {helpSections.map((section) => {
              const isOpen = openSections.includes(section.id);
              return (
                <div key={section.id} className="border rounded-lg">
                  <Button
                    variant="ghost"
                    className="w-full justify-between p-4 h-auto text-left"
                    onClick={() => toggleSection(section.id)}
                  >
                    <div className="flex items-center gap-2">
                      {section.icon}
                      <span className="font-medium">{section.title}</span>
                    </div>
                    {isOpen ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                  </Button>
                  {isOpen && (
                    <div className="px-4 pb-4 pt-2 border-t">
                      {section.content}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};
