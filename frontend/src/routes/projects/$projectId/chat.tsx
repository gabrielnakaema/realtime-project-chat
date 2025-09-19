import { AddProjectMember } from '@/components/add-project-member';
import { useAuth } from '@/hooks/use-auth';
import { createMessage, getChatByProjectId, listMessagesByProjectId } from '@/services/chat';
import { getProject } from '@/services/projects';
import { chatQueryKeys, projectQueryKeys } from '@/services/query-keys';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createFileRoute, Link } from '@tanstack/react-router';
import { ArrowLeft, Send, Users } from 'lucide-react';
import { useState } from 'react';

export const Route = createFileRoute('/projects/$projectId/chat')({
  component: RouteComponent,
});

function RouteComponent() {
  const { projectId } = Route.useParams();
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const [newMessage, setNewMessage] = useState('');
  const [before, setBefore] = useState<string | null>(null);

  const { data: project } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
  });

  const { data: chatData } = useQuery({
    queryKey: chatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const { data: messagesData } = useQuery({
    queryKey: chatQueryKeys.listMessagesByProjectId(projectId, before || ''),
    queryFn: () => listMessagesByProjectId(projectId, before || ''),
  });

  const { mutate } = useMutation({
    mutationFn: createMessage,
    onSuccess: () => {
      setNewMessage('');
      queryClient.invalidateQueries({ queryKey: chatQueryKeys.listMessagesByProjectId(projectId, before || '') });
    },
  });

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const formatDate = (timestamp: string) => {
    return new Date(timestamp).toLocaleDateString([], { month: 'short', day: 'numeric' });
  };

  const handleSendMessage = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!chatData?.id) {
      return;
    }

    mutate({
      chat_id: chatData.id,
      content: newMessage,
    });
  };

  const messages = messagesData?.data || [];

  return (
    <div className="flex flex-col h-screen flex-1">
      <header className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects/$projectId"
                className="inline-flex items-center px-3 py-2 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-md font-medium transition-colors"
                params={{ projectId }}
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Go back
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">Team chat - {project?.name}</h1>
                <p className="text-slate-600 dark:text-slate-400">Chat with your team members</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <AddProjectMember projectId={projectId} />
                <Users className="w-4 h-4 text-slate-500" />

                <div className="flex -space-x-2">
                  {project?.members.slice(0, 4).map((member) => (
                    <div
                      key={member.id}
                      className="w-8 h-8 bg-blue-600 rounded-full border-2 border-white dark:border-slate-800 flex items-center justify-center text-white text-xs font-medium"
                    >
                      {member?.user?.name.charAt(0).toUpperCase()}
                    </div>
                  ))}
                  {(project?.members?.length || 0) > 4 && (
                    <div className="w-8 h-8 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-white dark:border-slate-800 flex items-center justify-center">
                      <span className="text-xs font-medium text-slate-600 dark:text-slate-400">
                        +{(project?.members?.length || 0) - 4}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-4xl mx-auto space-y-4">
          {messages.map((message, index) => {
            const showDate =
              index === 0 || formatDate(message.created_at) !== formatDate(messages[index - 1].created_at);
            const isCurrentUser = message.member?.user?.id === user?.id;
            const isSystem = message.message_type === 'system';

            return (
              <div key={message.id}>
                {showDate && (
                  <div className="flex justify-center my-6">
                    <span className="bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-400 text-xs px-3 py-1 rounded-full">
                      {formatDate(message.created_at)}
                    </span>
                  </div>
                )}

                {isSystem ? (
                  <div className="flex justify-center">
                    <span className="text-sm text-slate-500 dark:text-slate-400 italic">{message.content}</span>
                  </div>
                ) : (
                  <div className={`flex gap-3 ${isCurrentUser ? 'flex-row-reverse' : ''}`}>
                    {!isCurrentUser && (
                      <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-xs font-medium mt-1">
                        {message.member?.user?.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <div className={`flex-1 max-w-md ${isCurrentUser ? 'text-right' : ''}`}>
                      {!isCurrentUser && (
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-sm font-medium text-slate-900 dark:text-slate-100">
                            {message.member?.user?.name}
                          </span>
                          <span className="text-xs text-slate-500 dark:text-slate-400">
                            {formatTime(message.created_at)}
                          </span>
                        </div>
                      )}
                      <div
                        className={`rounded-lg px-4 py-2 ${
                          isCurrentUser
                            ? 'bg-blue-600 text-white ml-auto'
                            : 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 border border-slate-200 dark:border-slate-700'
                        }`}
                      >
                        <p className="text-sm leading-relaxed">{message.content}</p>
                      </div>
                      {isCurrentUser && (
                        <div className="mt-1">
                          <span className="text-xs text-slate-500 dark:text-slate-400">
                            {formatTime(message.created_at)}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div className="border-t border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4">
        <div className="max-w-4xl mx-auto">
          <form onSubmit={handleSendMessage} className="flex items-end gap-3">
            <div className="flex-1">
              <div className="relative">
                <input
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  placeholder="Type your message..."
                  className="w-full px-3 py-3 pr-20 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 dark:placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
            <button
              type="submit"
              disabled={!newMessage.trim()}
              className="px-4 py-3 text-white bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 rounded-md font-medium transition-colors"
            >
              <Send className="w-4 h-4" />
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
