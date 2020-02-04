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

	SMFIA_INET    = '4'
	SMFIA_INET6   = '6'
	SMFIA_UNIX    = 'L'
	SMFIA_UNKNOWN = 'U'
)

const (
	SMFIS_KEEP    = uint32(20)
	SMFIS_ABORT   = uint32(21)
	SMFIS_OPTIONS = uint32(22)
	SMFIS_NOREPLY = uint32(7)
)

// Stage is for SetSymList, to tell the MTA what Macros in what Stage we need
type Stage uint32

const (
	SMFIM_CONNECT Stage = iota /* 0 connect */
	SMFIM_HELO                 /* 1 HELO/EHLO */
	SMFIM_ENVFROM              /* 2 MAIL From */
	SMFIM_ENVRCPT              /* 3 RCPT To */
	SMFIM_DATA                 /* 4 DATA */
	SMFIM_EOM                  /* 5 = end of message (final dot) */
	SMFIM_EOH                  /* 6 = end of header */
)

// OptAction sets which actions the milter wants to perform.
// Multiple options can be set using a bitmask.
type OptAction uint32

// OptProtocol masks out unwanted parts of the SMTP transaction.
// Multiple options can be set using a bitmask.
type OptProtocol uint32

const (
	// set which actions the milter wants to perform
	OptNone           OptAction = 0x00  /* SMFIF_NONE no flags */
	OptAddHeader      OptAction = 0x01  /* SMFIF_ADDHDRS filter may add headers */
	OptChangeBody     OptAction = 0x02  /* SMFIF_CHGBODY filter may replace body */
	OptAddRcpt        OptAction = 0x04  /* SMFIF_ADDRCPT filter may add recipients */
	OptRemoveRcpt     OptAction = 0x08  /* SMFIF_DELRCPT filter may delete recipients */
	OptChangeHeader   OptAction = 0x10  /* SMFIF_CHGHDRS filter may change/delete headers */
	OptQuarantine     OptAction = 0x20  /* SMFIF_QUARANTINE filter may quarantine envelope */
	OptChangeFrom     OptAction = 0x40  /* SMFIF_CHGFROM filter may change "from" (envelope sender) */
	OptAddRcptPartial OptAction = 0x80  /* SMFIF_ADDRCPT_PAR filter may add recipients, including ESMTP parameter to the envelope.*/
	OptSetSymList     OptAction = 0x100 /* SMFIF_SETSYMLIST filter can send set of symbols (macros) that it wants */
	// OptAllActions SMFI_CURR_ACTS Set of all actions in the current milter version */
	OptAllActions OptAction = OptAddHeader | OptChangeBody | OptAddRcpt | OptRemoveRcpt | OptChangeHeader | OptQuarantine | OptChangeFrom | OptAddRcptPartial | OptSetSymList

	// mask out unwanted parts of the SMTP transaction
	OptNoConnect    OptProtocol = 0x01       /* SMFIP_NOCONNECT MTA should not send connect info */
	OptNoHelo       OptProtocol = 0x02       /* SMFIP_NOHELO MTA should not send HELO info */
	OptNoMailFrom   OptProtocol = 0x04       /* SMFIP_NOMAIL MTA should not send MAIL info */
	OptNoRcptTo     OptProtocol = 0x08       /* SMFIP_NORCPT MTA should not send RCPT info */
	OptNoBody       OptProtocol = 0x10       /* SMFIP_NOBODY MTA should not send body (chunk) */
	OptNoHeaders    OptProtocol = 0x20       /* SMFIP_NOHDRS MTA should not send headers */
	OptNoEOH        OptProtocol = 0x40       /* SMFIP_NOEOH MTA should not send EOH */
	OptNrHdr        OptProtocol = 0x80       /* SMFIP_NR_HDR SMFIP_NOHREPL No reply for headers */
	OptNoUnknown    OptProtocol = 0x100      /* SMFIP_NOUNKNOWN MTA should not send unknown commands */
	OptNoData       OptProtocol = 0x200      /* SMFIP_NODATA MTA should not send DATA */
	OptSkip         OptProtocol = 0x400      /* SMFIP_SKIP MTA understands SMFIS_SKIP */
	OptRcptRej      OptProtocol = 0x800      /* SMFIP_RCPT_REJ MTA should also send rejected RCPTs */
	OptNrConn       OptProtocol = 0x1000     /* SMFIP_NR_CONN No reply for connect */
	OptNrHelo       OptProtocol = 0x2000     /* SMFIP_NR_HELO No reply for HELO */
	OptNrMailFrom   OptProtocol = 0x4000     /* SMFIP_NR_MAIL No reply for MAIL */
	OptNrRcptTo     OptProtocol = 0x8000     /* SMFIP_NR_RCPT No reply for RCPT */
	OptNrData       OptProtocol = 0x10000    /* SMFIP_NR_DATA No reply for DATA */
	OptNrUnknown    OptProtocol = 0x20000    /* SMFIP_NR_UNKN No reply for UNKNOWN */
	OptNrEOH        OptProtocol = 0x40000    /* SMFIP_NR_EOH No reply for eoh */
	OptNrBody       OptProtocol = 0x80000    /* SMFIP_NR_BODY No reply for body chunk */
	OptHdrLeadSpace OptProtocol = 0x100000   /* SMFIP_HDR_LEADSPC header value leading space */
	OptMDS256K      OptProtocol = 0x10000000 /* SMFIP_MDS_256K MILTER_MAX_DATA_SIZE=256K */
	OptMDS1M        OptProtocol = 0x20000000 /* SMFIP_MDS_1M MILTER_MAX_DATA_SIZE=1M */
)
