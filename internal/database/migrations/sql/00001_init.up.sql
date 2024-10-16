create table logs (
  version integer,
  time text,
  component text,
  message text,
  original text
);

create virtual table logs_fts using fts5(
  version unindexed,
  time unindexed,
  component,
  message,
  original,
  content='logs'
);

create trigger logs_after_insert after insert on logs
  begin
    insert into logs_fts (rowid, component, message, original)
    values (new.rowid, new.component, new.message, new.original);
  end;

create trigger logs_after_delete after delete on logs
  begin
    insert into logs_fts (logs_fts, rowid, component, message, original)
    values ('delete', new.rowid, new.component, new.message, new.original);
  end;

create trigger logs_after_update after update on logs
  begin
    insert into logs_fts (logs_fts, rowid, component, message, original)
    values ('delete', new.rowid, new.component, new.message, new.original);
    insert into logs_fts (rowid, component, message, original)
    values (new.rowid, new.component, new.message, new.original);
  end;
