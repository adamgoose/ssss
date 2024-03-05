DEFINE TABLE secrets SCHEMAFULL;

DEFINE FIELD user ON secrets TYPE record<users>;
DEFINE FIELD label ON secrets TYPE string;
DEFINE FIELD parts ON secrets TYPE int;
DEFINE FIELD threshold ON secrets TYPE int;
DEFINE FIELD status ON secrets TYPE string;
DEFINE FIELD created_at ON secrets TYPE datetime;
