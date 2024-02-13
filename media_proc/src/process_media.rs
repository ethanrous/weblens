use blurhash::encode;
use image::{GenericImageView, EncodableLayout, image, load_from_memory};

fn generate_blurhash(i: image) {
    let (width, height) = img.dimensions();
    let blurhash = encode(4, 3, width, height, img.to_rgba8().as_bytes()).unwrap();
}

pub fn bytes_to_image(bytes: &mut [u8]) {
    img = image::load_from_memory(bytes);
    reutrn;
}