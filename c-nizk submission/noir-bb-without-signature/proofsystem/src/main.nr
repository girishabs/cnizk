use std::hash::blake2s;

// -------------------------------
// Public constraint struct
// -------------------------------
struct Constraint {
    pub operation: u32,
    pub value: u32,
    pub attribute_type: u32,
}

// -------------------------------
// Main circuit
// -------------------------------
fn main(
    data: [u32; 5],
    creation_timestamp: pub u32,
    randomness: [u8; 16],
    constraints: pub [Constraint; 5],
    message: pub [u8; 32],
) {
    // ---------------------------------
    // Validate constraints
    // ---------------------------------
    for i in 0..5 {
        let d = data[i];
        let op = constraints[i].operation;
        let val = constraints[i].value;
        let t = constraints[i].attribute_type;

        // op in {0,1,2,3,4,5}
        if op > 5 {
            assert(false);
        }

        // string is only eq / ne
        if t == 1 {
            if op != 0 {
                assert(op == 1);
            }
        }

        // bool in {0,1}
        if t == 2 {
            assert(d * (d - 1) == 0);
            assert(val * (val - 1) == 0);
        }

        // Apply constraint
        if op == 0 { assert(d == val); }  // EQ
        if op == 1 { assert(d != val); }  // NE
        if op == 2 { assert(d <  val); }  // LT
        if op == 3 { assert(d <= val); }  // LE
        if op == 4 { assert(d >  val); }  // GT
        if op == 5 { assert(d >= val); }  // GE

        assert(creation_timestamp == creation_timestamp)
    }

    let mut payload: [u8; 40] = [0; 40];

    for i in 0..5 {
    let v = data[i];

    payload[i * 4]     = ((v >> 24) & 0xff) as u8;
    payload[i * 4 + 1] = ((v >> 16) & 0xff) as u8;
    payload[i * 4 + 2] = ((v >> 8)  & 0xff) as u8;
    payload[i * 4 + 3] = (v & 0xff) as u8;
}

// timestamp (big-endian)
let ts = creation_timestamp;

payload[20] = ((ts >> 24) & 0xff) as u8;
payload[21] = ((ts >> 16) & 0xff) as u8;
payload[22] = ((ts >> 8)  & 0xff) as u8;
payload[23] = (ts & 0xff) as u8;

for i in 0..16 {
    payload[24 + i] = randomness[i];
}
// ---------------------------------
// Blake2s commitment
// ---------------------------------

let computed_commitment = blake2s(payload);

assert(computed_commitment == message);
}