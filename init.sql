/* gateway */

DROP TABLE IF EXISTS gw_nodes;
DROP TABLE IF EXISTS gw_url;
DROP TABLE IF EXISTS gw_blocklist;

CREATE TABLE gw_nodes (
	address  VARCHAR(255) NOT NULL,
	outbound BOOL,
	PRIMARY KEY (address)
);

CREATE TABLE gw_url (
	router_url VARCHAR(255) NOT NULL
);

INSERT INTO gw_url (router_url) VALUES ('');

CREATE TABLE gw_blocklist (
	ip VARCHAR(255) NOT NULL,
	PRIMARY KEY (ip)
);

/* consensus */

DROP TABLE IF EXISTS cs_height;
DROP TABLE IF EXISTS cs_consistency;
DROP TABLE IF EXISTS cs_sfpool;
DROP TABLE IF EXISTS cs_changelog;
DROP TABLE IF EXISTS cs_dsco;
DROP TABLE IF EXISTS cs_fcex;
DROP TABLE IF EXISTS cs_oak;
DROP TABLE IF EXISTS cs_oak_init;
DROP TABLE IF EXISTS cs_sco;
DROP TABLE IF EXISTS cs_fc;
DROP TABLE IF EXISTS cs_sfo;
DROP TABLE IF EXISTS cs_fuh;
DROP TABLE IF EXISTS cs_fuh_current;
DROP TABLE IF EXISTS cs_map;
DROP TABLE IF EXISTS cs_path;
DROP TABLE IF EXISTS cs_cl;
DROP TABLE IF EXISTS cs_dos;

CREATE TABLE cs_height (
	id     INT NOT NULL AUTO_INCREMENT,
	height BIGINT UNSIGNED NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_consistency (
	id            INT NOT NULL AUTO_INCREMENT,
	inconsistency BOOL NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_sfpool (
	id    INT NOT NULL AUTO_INCREMENT,
	bytes VARBINARY(24) NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_changelog (
	id    INT NOT NULL AUTO_INCREMENT,
	bytes BINARY(32) NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_dsco (
	height BIGINT UNSIGNED NOT NULL,
	scoid  BINARY(32) NOT NULL,
	bytes  VARBINARY(56) NOT NULL,
	PRIMARY KEY (scoid ASC)
);

CREATE TABLE cs_fcex (
	height BIGINT UNSIGNED NOT NULL,
	fcid   BINARY(32) NOT NULL,
	bytes  BLOB NOT NULL,
	PRIMARY KEY (fcid ASC)
);

CREATE TABLE cs_oak (
	bid   BINARY(32) NOT NULL UNIQUE,
	bytes BINARY(40) NOT NULL,
	PRIMARY KEY (bid ASC)
);

CREATE TABLE cs_oak_init (
	id   INT NOT NULL AUTO_INCREMENT,
	init BOOL NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_sco (
	scoid BINARY(32) NOT NULL,
	bytes VARBINARY(56) NOT NULL,
	PRIMARY KEY (scoid ASC)
);

CREATE TABLE cs_fc (
	fcid  BINARY(32) NOT NULL,
	bytes BLOB NOT NULL,
	PRIMARY KEY (fcid ASC)
);

CREATE TABLE cs_sfo (
	sfoid BINARY(32) NOT NULL,
	bytes VARBINARY(80) NOT NULL,
	PRIMARY KEY (sfoid ASC)
);

CREATE TABLE cs_fuh (
	height BIGINT UNSIGNED NOT NULL,
	bytes  BINARY(64) NOT NULL,
	PRIMARY KEY (height ASC)
);

CREATE TABLE cs_fuh_current (
	id     INT NOT NULL AUTO_INCREMENT,
	bytes  BINARY(64) NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_path (
	height BIGINT UNSIGNED NOT NULL,
	bid    BINARY(32) NOT NULL,
	PRIMARY KEY (height ASC)
);

CREATE TABLE cs_map (
	id    INT NOT NULL AUTO_INCREMENT,
	bid   BINARY(32) NOT NULL UNIQUE,
	bytes LONGBLOB NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE cs_cl (
	ceid  BINARY(32) NOT NULL,
	bytes VARBINARY(1024) NOT NULL,
	PRIMARY KEY (ceid ASC)
);

CREATE TABLE cs_dos (
	bid BINARY(32) NOT NULL,
	PRIMARY KEY (bid ASC)
);

/* transactionpool */

DROP TABLE IF EXISTS tp_height;
DROP TABLE IF EXISTS tp_ctx;
DROP TABLE IF EXISTS tp_median;
DROP TABLE IF EXISTS tp_cc;
DROP TABLE IF EXISTS tp_recent;

CREATE TABLE tp_height (
	id     INT NOT NULL AUTO_INCREMENT,
	height BIGINT UNSIGNED NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE tp_ctx (
	txid BINARY(32) NOT NULL,
	PRIMARY KEY (txid),
	INDEX txid (txid ASC)
);

CREATE TABLE tp_median (
	id    INT NOT NULL AUTO_INCREMENT,
	bytes BLOB NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE tp_cc (
	id   INT NOT NULL AUTO_INCREMENT,
	ceid BINARY(32) NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE tp_recent (
	id  INT NOT NULL AUTO_INCREMENT,
	bid BINARY(32) NOT NULL,
	PRIMARY KEY (id)
);

/* wallet */

DROP TABLE IF EXISTS wt_addr;
DROP TABLE IF EXISTS wt_txn;
DROP TABLE IF EXISTS wt_sco;
DROP TABLE IF EXISTS wt_sfo;
DROP TABLE IF EXISTS wt_spo;
DROP TABLE IF EXISTS wt_uc;
DROP TABLE IF EXISTS wt_info;
DROP TABLE IF EXISTS wt_watch;
DROP TABLE IF EXISTS wt_aux;
DROP TABLE IF EXISTS wt_keys;

CREATE TABLE wt_txn (
	id    INT NOT NULL AUTO_INCREMENT,
	txid  BINARY(32) NOT NULL UNIQUE,
	bytes BLOB NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE wt_addr (
	id   INT NOT NULL AUTO_INCREMENT,
	addr BINARY(32) NOT NULL,
	txid BINARY(32) NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (txid) REFERENCES wt_txn(txid)
);

CREATE TABLE wt_sco (
	scoid BINARY(32) NOT NULL,
	bytes VARBINARY(56) NOT NULL,
	PRIMARY KEY (scoid ASC)
);

CREATE TABLE wt_sfo (
	sfoid BINARY(32) NOT NULL,
	bytes VARBINARY(80) NOT NULL,
	PRIMARY KEY (sfoid ASC)
);

CREATE TABLE wt_spo (
	oid    BINARY(32) NOT NULL,
	height BIGINT UNSIGNED NOT NULL,
	PRIMARY KEY (oid ASC)
);

CREATE TABLE wt_uc (
	addr  BINARY(32) NOT NULL,
	bytes BLOB NOT NULL,
	PRIMARY KEY (addr ASC)
);

CREATE TABLE wt_info (
	id        INT NOT NULL AUTO_INCREMENT,
	cc        BINARY(32) NOT NULL,
	height    BIGINT UNSIGNED NOT NULL,
	encrypted BLOB NOT NULL,
	sfpool    VARBINARY(24) NOT NULL,
	salt      BINARY(32) NOT NULL,
	progress  BIGINT UNSIGNED NOT NULL,
	seed      BLOB NOT NULL,
	pwd       BLOB NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE wt_aux (
	salt      BINARY(32) NOT NULL,
	encrypted BLOB NOT NULL,
	seed      BLOB NOT NULL,
	PRIMARY KEY (seed(32))
);

CREATE TABLE wt_keys (
	salt      BINARY(32) NOT NULL,
	encrypted BLOB NOT NULL,
	skey      BLOB NOT NULL,
	PRIMARY KEY (skey(32))
);

CREATE TABLE wt_watch (
	addr BINARY(32) NOT NULL,
	PRIMARY KEY (addr ASC)
);

/* manager */

DROP TABLE IF EXISTS mg_timestamp;

CREATE TABLE mg_timestamp (
	id     INT NOT NULL AUTO_INCREMENT,
	height BIGINT UNSIGNED NOT NULL,
	time   BIGINT UNSIGNED NOT NULL,
	PRIMARY KEY (id)
);

/* hostdb */

DROP TABLE IF EXISTS hdb_scanhistory;
DROP TABLE IF EXISTS hdb_ipnets;
DROP TABLE IF EXISTS hdb_hosts;
DROP TABLE IF EXISTS hdb_fdomains;
DROP TABLE IF EXISTS hdb_fhosts;
DROP TABLE IF EXISTS hdb_contracts;
DROP TABLE IF EXISTS hdb_info;

CREATE TABLE hdb_hosts (
	public_key BINARY(32) NOT NULL,
	filtered   BOOL NOT NULL,
	bytes      BLOB NOT NULL,
	PRIMARY KEY (public_key)
);

CREATE TABLE hdb_scanhistory (
	public_key BINARY(32) NOT NULL,
	time       BIGINT UNSIGNED NOT NULL,
	success    BOOL NOT NULL
);

CREATE TABLE hdb_ipnets (
	public_key BINARY(32) NOT NULL,
	ip_net     VARCHAR(255) NOT NULL
);

CREATE TABLE hdb_fdomains (
	dom VARCHAR(255) NOT NULL
);

CREATE TABLE hdb_fhosts (
	public_key BINARY(32) NOT NULL
);

CREATE TABLE hdb_contracts (
	host_pk   BINARY(32) NOT NULL,
	renter_pk BINARY(32) NOT NULL,
	data      BIGINT UNSIGNED NOT NULL
);

CREATE TABLE hdb_info (
	id               INT NOT NULL AUTO_INCREMENT,
	height           BIGINT UNSIGNED NOT NULL,
	scan_complete    BOOL NOT NULL,
	disable_ip_check BOOL NOT NULL,
	last_change      BINARY(32) NOT NULL,
	filter_mode      INT NOT NULL,
	PRIMARY KEY (id)
);

/* satellite */

DROP TABLE IF EXISTS spendings;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS renters;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS contracts;
DROP TABLE IF EXISTS accounts;

CREATE TABLE accounts (
	id             INT NOT NULL AUTO_INCREMENT,
	email          VARCHAR(64) NOT NULL UNIQUE,
	password_hash  VARCHAR(64) NOT NULL,
	verified       BOOL NOT NULL,
	created        INT NOT NULL,
	nonce          VARCHAR(32) NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE balances (
	id         INT NOT NULL AUTO_INCREMENT,
	email      VARCHAR(64) NOT NULL UNIQUE,
	subscribed BOOL NOT NULL,
	balance    DOUBLE NOT NULL,
	locked     DOUBLE NOT NULL,
	currency   VARCHAR(8) NOT NULL,
	stripe_id  VARCHAR(32) NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (email) REFERENCES accounts(email)
);

CREATE TABLE payments (
	id        INT NOT NULL AUTO_INCREMENT,
	email     VARCHAR(64) NOT NULL,
	amount    DOUBLE NOT NULL,
	currency  VARCHAR(8) NOT NULL,
	amount_sc DOUBLE NOT NULL,
	made_at   INT NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (email) REFERENCES accounts(email)
);

CREATE TABLE spendings (
	id               INT NOT NULL AUTO_INCREMENT,
	email            VARCHAR(64) NOT NULL UNIQUE,
	current_locked   DOUBLE NOT NULL,
	current_used     DOUBLE NOT NULL,
	current_overhead DOUBLE NOT NULL,
	prev_locked      DOUBLE NOT NULL,
	prev_used        DOUBLE NOT NULL,
	prev_overhead    DOUBLE NOT NULL,
	current_formed   BIGINT UNSIGNED NOT NULL,
	current_renewed  BIGINT UNSIGNED NOT NULL,
	prev_formed      BIGINT UNSIGNED NOT NULL,
	prev_renewed     BIGINT UNSIGNED NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (email) REFERENCES accounts(email)
);

CREATE TABLE renters (
	id                           INT NOT NULL AUTO_INCREMENT,
	email                        VARCHAR(64) NOT NULL UNIQUE,
	public_key                   VARCHAR(128) NOT NULL UNIQUE,
	current_period               BIGINT UNSIGNED NOT NULL,
	funds                        VARCHAR(64) NOT NULL,
	hosts                        BIGINT UNSIGNED NOT NULL,
	period                       BIGINT UNSIGNED NOT NULL,
	renew_window                 BIGINT UNSIGNED NOT NULL,
	expected_storage             BIGINT UNSIGNED NOT NULL,
	expected_upload              BIGINT UNSIGNED NOT NULL,
	expected_download            BIGINT UNSIGNED NOT NULL,
	min_shards                   BIGINT UNSIGNED NOT NULL,
	total_shards                 BIGINT UNSIGNED NOT NULL,
	max_rpc_price                VARCHAR(64) NOT NULL,
	max_contract_price           VARCHAR(64) NOT NULL,
	max_download_bandwidth_price VARCHAR(64) NOT NULL,
	max_sector_access_price      VARCHAR(64) NOT NULL,
	max_storage_price            VARCHAR(64) NOT NULL,
	max_upload_bandwidth_price   VARCHAR(64) NOT NULL,
	min_max_collateral           VARCHAR(64) NOT NULL,
	blockheight_leeway           BIGINT UNSIGNED NOT NULL,
	private_key                  VARCHAR(128) NOT NULL,
	auto_renew_contracts         BOOL NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (email) REFERENCES accounts(email)
);

CREATE TABLE contracts (
	id                      INT NOT NULL AUTO_INCREMENT,
	contract_id             VARCHAR(64) NOT NULL UNIQUE,
	renter_pk               VARCHAR(128) NOT NULL,
	start_height            BIGINT UNSIGNED NOT NULL,
	download_spending       VARCHAR(64) NOT NULL,
	fund_account_spending   VARCHAR(64) NOT NULL,
	storage_spending        VARCHAR(64) NOT NULL,
	upload_spending         VARCHAR(64) NOT NULL,
	total_cost              VARCHAR(64) NOT NULL,
	contract_fee            VARCHAR(64) NOT NULL,
	txn_fee                 VARCHAR(64) NOT NULL,
	siafund_fee             VARCHAR(64) NOT NULL,
	account_balance_cost    VARCHAR(64) NOT NULL,
	fund_account_cost       VARCHAR(64) NOT NULL,
	update_price_table_cost VARCHAR(64) NOT NULL,
	good_for_upload         BOOL NOT NULL,
	good_for_renew          BOOL NOT NULL,
	bad_contract            BOOL NOT NULL,
	last_oos_err            BIGINT UNSIGNED NOT NULL,
	locked                  BOOL NOT NULL,
	renewed_from            VARCHAR(64) NOT NULL,
	renewed_to              VARCHAR(64) NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE transactions (
	id                           INT NOT NULL AUTO_INCREMENT,
	contract_id                  VARCHAR(64) NOT NULL UNIQUE,
	parent_id                    VARCHAR(64) NOT NULL,
	uc_timelock                  BIGINT UNSIGNED NOT NULL,
	uc_renter_pk                 VARCHAR(128) NOT NULL,
	uc_host_pk                   VARCHAR(128) NOT NULL,
	signatures_required          INT NOT NULL,
	new_revision_number          BIGINT UNSIGNED NOT NULL,
	new_file_size                BIGINT UNSIGNED NOT NULL,
	new_file_merkle_root         VARCHAR(64) NOT NULL,
	new_window_start             BIGINT UNSIGNED NOT NULL,
	new_window_end               BIGINT UNSIGNED NOT NULL,
	new_valid_proof_output_0     VARCHAR(64) NOT NULL,
	new_valid_proof_output_uh_0  VARCHAR(64) NOT NULL,
	new_valid_proof_output_1     VARCHAR(64) NOT NULL,
	new_valid_proof_output_uh_1  VARCHAR(64) NOT NULL,
	new_missed_proof_output_0    VARCHAR(64) NOT NULL,
	new_missed_proof_output_uh_0 VARCHAR(64) NOT NULL,
	new_missed_proof_output_1    VARCHAR(64) NOT NULL,
	new_missed_proof_output_uh_1 VARCHAR(64) NOT NULL,
	new_missed_proof_output_2    VARCHAR(64) NOT NULL,
	new_missed_proof_output_uh_2 VARCHAR(64) NOT NULL,
	new_unlock_hash              VARCHAR(64) NOT NULL,
	t_parent_id_0                VARCHAR(64) NOT NULL,
	pk_index_0                   BIGINT UNSIGNED NOT NULL,
	timelock_0                   BIGINT UNSIGNED NOT NULL,
	signature_0                  VARCHAR(128) NOT NULL,
	t_parent_id_1                VARCHAR(64) NOT NULL,
	pk_index_1                   BIGINT UNSIGNED NOT NULL,
	timelock_1                   BIGINT UNSIGNED NOT NULL,
	signature_1                  VARCHAR(128) NOT NULL,
	PRIMARY KEY (id),
	FOREIGN KEY (contract_id) REFERENCES contracts(contract_id)
);
