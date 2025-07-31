-- name: CountContactsAndJournalEntries :one
select (
        select count(*)
        from contacts
        where contacts.namespace = $1
    ) as contact_count,
    (
        select count(*)
        from journal_entries
        where journal_entries.namespace = $1
    ) as journal_entries_count;