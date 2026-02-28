use std::fs;
use std::io::{self, Cursor, Read};

/// Parsed header fields from an IRG file.
#[derive(Debug)]
pub struct IrgHeader {
    pub unknown_header: i32,
    pub first_image_size: u32,
    pub first_image_width: u16,
    pub first_image_height: u16,
    pub second_image_size: u32,
    pub second_image_width: u16,
    pub second_image_height: u16,
    pub third_image_size: u32,
    pub third_image_width: u16,
    pub third_image_height: u16,
    pub emissivity: u32,
    pub reflective_temperature: u32,
    pub ambient_temperature: u32,
    pub distance: u32,
    pub transmissivity: u16,
}

/// Extracted image data from an IRG file.
pub struct IrgData {
    pub header: IrgHeader,
    pub grayscale: Vec<u8>,
    pub thermal_raw: Vec<u16>,
    pub thermal_width: usize,
    pub thermal_height: usize,
}

impl IrgData {
    /// Return thermal data converted to Fahrenheit.
    pub fn thermal_fahrenheit(&self) -> Vec<f64> {
        self.thermal_raw
            .iter()
            .map(|&v| {
                let kelvin = f64::from(v) / 10.0;
                (kelvin - 273.15) * 9.0 / 5.0 + 32.0
            })
            .collect()
    }
}

/// Read a little-endian u16 from a cursor.
fn read_u16_le(cur: &mut Cursor<&[u8]>) -> io::Result<u16> {
    let mut buf = [0u8; 2];
    cur.read_exact(&mut buf)?;
    Ok(u16::from_le_bytes(buf))
}

/// Read a little-endian u32 from a cursor.
fn read_u32_le(cur: &mut Cursor<&[u8]>) -> io::Result<u32> {
    let mut buf = [0u8; 4];
    cur.read_exact(&mut buf)?;
    Ok(u32::from_le_bytes(buf))
}

/// Read a little-endian i32 from a cursor.
fn read_i32_le(cur: &mut Cursor<&[u8]>) -> io::Result<i32> {
    let mut buf = [0u8; 4];
    cur.read_exact(&mut buf)?;
    Ok(i32::from_le_bytes(buf))
}

/// Parse an IRG file and extract all image data.
pub fn extract_irg(path: &str) -> io::Result<IrgData> {
    let bytes = fs::read(path)?;
    if bytes.len() < 0x100 {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "File too small to be a valid IRG file",
        ));
    }

    let mut cur = Cursor::new(bytes.as_slice());

    // Parse header (matches the Python struct format '<iIHHbIHHBIHHIIIIIIIH14xB')
    let unknown_header = read_i32_le(&mut cur)?;
    let first_image_size = read_u32_le(&mut cur)?;
    let first_image_width = read_u16_le(&mut cur)?;
    let first_image_height = read_u16_le(&mut cur)?;
    // pad1 (1 byte)
    let mut _pad = [0u8; 1];
    cur.read_exact(&mut _pad)?;

    let second_image_size = read_u32_le(&mut cur)?;
    let second_image_width = read_u16_le(&mut cur)?;
    let second_image_height = read_u16_le(&mut cur)?;
    // pad2 (1 byte)
    cur.read_exact(&mut _pad)?;

    let third_image_size = read_u32_le(&mut cur)?;
    let third_image_width = read_u16_le(&mut cur)?;
    let third_image_height = read_u16_le(&mut cur)?;

    let emissivity = read_u32_le(&mut cur)?;
    let reflective_temperature = read_u32_le(&mut cur)?;
    let ambient_temperature = read_u32_le(&mut cur)?;
    let distance = read_u32_le(&mut cur)?;
    let _unknown = read_u32_le(&mut cur)?;
    let _transmissivity_u32 = read_u32_le(&mut cur)?; // padding/unknown
    let _padding = read_u32_le(&mut cur)?;
    let transmissivity = read_u16_le(&mut cur)?;

    let header = IrgHeader {
        unknown_header,
        first_image_size,
        first_image_width,
        first_image_height,
        second_image_size,
        second_image_width,
        second_image_height,
        third_image_size,
        third_image_width,
        third_image_height,
        emissivity,
        reflective_temperature,
        ambient_temperature,
        distance,
        transmissivity,
    };

    // Determine data start offset
    let data_start = if bytes[0x7E] == 0xAC && bytes[0x7F] == 0xCA {
        0x80usize
    } else {
        0x100usize
    };

    let mut offset = data_start;

    // Extract grayscale image (uint8 array)
    let gs_size = first_image_size as usize;
    if offset + gs_size > bytes.len() {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "File too small for grayscale data",
        ));
    }
    let grayscale = bytes[offset..offset + gs_size].to_vec();
    offset += gs_size;

    // Extract thermal data (uint16 array, 2 bytes per element)
    let thermal_count = second_image_size as usize;
    let thermal_byte_size = thermal_count * 2;
    if offset + thermal_byte_size > bytes.len() {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "File too small for thermal data",
        ));
    }
    let thermal_raw: Vec<u16> = bytes[offset..offset + thermal_byte_size]
        .chunks_exact(2)
        .map(|c| u16::from_le_bytes([c[0], c[1]]))
        .collect();

    Ok(IrgData {
        header,
        grayscale,
        thermal_raw,
        thermal_width: second_image_width as usize,
        thermal_height: second_image_height as usize,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_read_u16_le() {
        let data = [0x34, 0x12];
        let mut cur = Cursor::new(data.as_slice());
        assert_eq!(read_u16_le(&mut cur).unwrap(), 0x1234);
    }

    #[test]
    fn test_read_u32_le() {
        let data = [0x78, 0x56, 0x34, 0x12];
        let mut cur = Cursor::new(data.as_slice());
        assert_eq!(read_u32_le(&mut cur).unwrap(), 0x12345678);
    }
}
