# Decoder compatibility policy

Record sizes count UTF-8 bytes excluding LF or CRLF separators. Whitespace-only
records are ignored. The compatibility default is 1 MiB per record; explicit
limits must be positive and no greater than 16 MiB. Reject invalid limits before
reading input. The maximum permitted size itself is accepted.

Errors must support errors.Is with exported sentinel values ErrInvalidLimit,
ErrRecordTooLarge and ErrInvalidRecord. Never include record content in error
messages. Preserve reader I/O errors through wrapping. On any error return no
records, so callers cannot accidentally accept a partial batch. Do not allocate
the entire declared limit before reading a record.
