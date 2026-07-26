--liquibase formatted sql

--changeset pixels:chat-0002-add-message-search-index
create index chat_messages_message_search_idx
on chat_messages
using gin (to_tsvector('simple', message));

--rollback drop index if exists chat_messages_message_search_idx;
