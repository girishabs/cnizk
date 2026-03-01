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
    signature: [u8; 64],

    constraints: pub [Constraint; 5],
    creation_timestamp: pub u32,
    message: pub [u8; 32],

    pub_key_x: pub [u8; 32],
    pub_key_y: pub [u8; 32],
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
    }


let mut payload: [u8; 24] = [0; 24];

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

// ---------------------------------
// Blake2s commitment
// ---------------------------------

let computed_commitment = blake2s(payload);

assert(computed_commitment == message);

    // ---------------------------------
    // ECDSA secp256k1 signature check
    // ---------------------------------

    // ---------------------------------
    // ECDSA secp256k1 signature check
    // ---------------------------------

    let ok = std::ecdsa_secp256k1::verify_signature(pub_key_x, pub_key_y, signature, message);
    assert(ok);
}