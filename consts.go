package milter

const (
	SMFIC_ABORT   = 'A' // Abort current filter checks
	SMFIC_BODY    = 'B' // Body chunk
	SMFIC_CONNECT = 'C' // SMTP connection information
	SMFIC_MACRO   = 'D' // Define macros
	SMFIC_BODYEOB = 'E' // End of body marker
	SMFIC_HELO    = 'H' // HELO/EHLO name
	SMFIC_QUIT_NC = 'K' // QUIT but new connection follows
	SMFIC_HEADER  = 'L' // Mail header
	SMFIC_MAIL    = 'M' // MAIL FROM: information
	SMFIC_EOH     = 'N' // End of headers marker
	SMFIC_OPTNEG  = 'O' // Option negotiation
	SMFIC_QUIT    = 'Q' // Quit milter communication
	SMFIC_RCPT    = 'R' // RCPT TO: information
	SMFIC_DATA    = 'T' // DATA
	SMFIC_UNKNOWN = 'U' // Any unknown command

	SMFIR_ADDRCPT     = '+' // Add recipient (modification action)
	SMFIR_DELRCPT     = '-' // Remove recipient (modification action)
	SMFIR_ADDRCPT_PAR = '2' // Add recipient (incl. ESMTP args)
	SMFIR_SHUTDOWN    = '4' // 421: shutdown (internal to MTA)
	SMFIR_ACCEPT      = 'a' // Accept message completely (accept/reject action)
	SMFIR_REPLBODY    = 'b' // Replace body (modification action)
	SMFIR_CONTINUE    = 'c' // Accept and keep processing (accept/reject action)
	SMFIR_DISCARD     = 'd' // Set discard flag for entire message (accept/reject action)
	SMFIR_CHGFROM     = 'e' // Change envelope sender (from)
	SMFIR_CONN_FAIL   = 'f' // Cause a connection failure
	SMFIR_ADDHEADER   = 'h' // Add header (modification action)
	SMFIR_INSHEADER   = 'i' // Insert header
	SMFIR_SETSYMLIST  = 'l' // Set list of symbols (macros)
	SMFIR_CHGHEADER   = 'm' // Change header (modification action)
	SMFIR_PROGRESS    = 'p' // Progress (asynchronous action)
	SMFIR_QUARANTINE  = 'q' // Quarantine message (modification action)
	SMFIR_REJECT      = 'r' // Reject command/recipient with a 5xx (accept/reject action)
	SMFIR_SKIP        = 's' // Skip
	SMFIR_TEMPFAIL    = 't' // Reject command/recipient with a 4xx (accept/reject action)
	SMFIR_REPLYCODE   = 'y' // Send specific Nxx reply message (accept/reject action)

	SMFIA_INET  = '4'
	SMFIA_INET6 = '6'
)
