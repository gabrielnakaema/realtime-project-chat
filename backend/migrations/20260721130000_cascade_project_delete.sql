-- +goose Up
-- +goose StatementBegin
ALTER TABLE project_members DROP CONSTRAINT fk_project_members_projects;
ALTER TABLE project_members ADD CONSTRAINT fk_project_members_projects FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE tasks DROP CONSTRAINT fk_tasks_projects;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_projects FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE tasks DROP CONSTRAINT fk_tasks_project_columns;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_project_columns FOREIGN KEY (project_column_id) REFERENCES project_columns(id) ON DELETE CASCADE;

ALTER TABLE project_activity_logs DROP CONSTRAINT fk_project_activity_logs_projects;
ALTER TABLE project_activity_logs ADD CONSTRAINT fk_project_activity_logs_projects FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE chats DROP CONSTRAINT fk_chats_projects;
ALTER TABLE chats ADD CONSTRAINT fk_chats_projects FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE chat_members DROP CONSTRAINT fk_chat_members_chats;
ALTER TABLE chat_members ADD CONSTRAINT fk_chat_members_chats FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE;

ALTER TABLE chat_messages DROP CONSTRAINT fk_chat_messages_chats;
ALTER TABLE chat_messages ADD CONSTRAINT fk_chat_messages_chats FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE;

ALTER TABLE task_tags DROP CONSTRAINT fk_task_tags_tasks;
ALTER TABLE task_tags ADD CONSTRAINT fk_task_tags_tasks FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;

ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_tasks;
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_tasks FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;

ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_parent_comments;
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_parent_comments FOREIGN KEY (parent_comment_id) REFERENCES task_comments(id) ON DELETE CASCADE;

ALTER TABLE task_changes DROP CONSTRAINT fk_task_changes_task_updates;
ALTER TABLE task_changes ADD CONSTRAINT fk_task_changes_task_updates FOREIGN KEY (update_id) REFERENCES task_updates(id) ON DELETE CASCADE;

ALTER TABLE notifications DROP CONSTRAINT fk_notifications_projects;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_projects FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE notifications DROP CONSTRAINT fk_notifications_tasks;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_tasks FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE;

ALTER TABLE notifications DROP CONSTRAINT fk_notifications_task_comments;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_task_comments FOREIGN KEY (task_comment_id) REFERENCES task_comments(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE project_members DROP CONSTRAINT fk_project_members_projects;
ALTER TABLE project_members ADD CONSTRAINT fk_project_members_projects FOREIGN KEY (project_id) REFERENCES projects(id);

ALTER TABLE tasks DROP CONSTRAINT fk_tasks_projects;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_projects FOREIGN KEY (project_id) REFERENCES projects(id);

ALTER TABLE tasks DROP CONSTRAINT fk_tasks_project_columns;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_project_columns FOREIGN KEY (project_column_id) REFERENCES project_columns(id);

ALTER TABLE project_activity_logs DROP CONSTRAINT fk_project_activity_logs_projects;
ALTER TABLE project_activity_logs ADD CONSTRAINT fk_project_activity_logs_projects FOREIGN KEY (project_id) REFERENCES projects(id);

ALTER TABLE chats DROP CONSTRAINT fk_chats_projects;
ALTER TABLE chats ADD CONSTRAINT fk_chats_projects FOREIGN KEY (project_id) REFERENCES projects(id);

ALTER TABLE chat_members DROP CONSTRAINT fk_chat_members_chats;
ALTER TABLE chat_members ADD CONSTRAINT fk_chat_members_chats FOREIGN KEY (chat_id) REFERENCES chats(id);

ALTER TABLE chat_messages DROP CONSTRAINT fk_chat_messages_chats;
ALTER TABLE chat_messages ADD CONSTRAINT fk_chat_messages_chats FOREIGN KEY (chat_id) REFERENCES chats(id);

ALTER TABLE task_tags DROP CONSTRAINT fk_task_tags_tasks;
ALTER TABLE task_tags ADD CONSTRAINT fk_task_tags_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);

ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_tasks;
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);

ALTER TABLE task_comments DROP CONSTRAINT fk_task_comments_parent_comments;
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_parent_comments FOREIGN KEY (parent_comment_id) REFERENCES task_comments(id);

ALTER TABLE task_changes DROP CONSTRAINT fk_task_changes_task_updates;
ALTER TABLE task_changes ADD CONSTRAINT fk_task_changes_task_updates FOREIGN KEY (update_id) REFERENCES task_updates(id);

ALTER TABLE notifications DROP CONSTRAINT fk_notifications_projects;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_projects FOREIGN KEY (project_id) REFERENCES projects(id);

ALTER TABLE notifications DROP CONSTRAINT fk_notifications_tasks;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_tasks FOREIGN KEY (task_id) REFERENCES tasks(id);

ALTER TABLE notifications DROP CONSTRAINT fk_notifications_task_comments;
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_task_comments FOREIGN KEY (task_comment_id) REFERENCES task_comments(id);
-- +goose StatementEnd
