mod irg;

use eframe::egui;
use irg::IrgData;

/// Apply an "inferno"-style colormap to a normalized value in [0, 1].
/// Returns (r, g, b) each in 0..255.
fn inferno_color(t: f64) -> [u8; 3] {
    // Simplified inferno colormap using key control points
    let t = t.clamp(0.0, 1.0);
    let (r, g, b) = if t < 0.25 {
        let s = t / 0.25;
        (
            0.0 + s * 0.34,
            0.0 + s * 0.06,
            0.01 + s * 0.33,
        )
    } else if t < 0.5 {
        let s = (t - 0.25) / 0.25;
        (
            0.34 + s * 0.48,
            0.06 + s * 0.06,
            0.34 + s * (-0.04),
        )
    } else if t < 0.75 {
        let s = (t - 0.5) / 0.25;
        (
            0.82 + s * 0.13,
            0.12 + s * 0.53,
            0.30 + s * (-0.26),
        )
    } else {
        let s = (t - 0.75) / 0.25;
        (
            0.95 + s * 0.03,
            0.65 + s * 0.30,
            0.04 + s * 0.90,
        )
    };
    [
        (r.clamp(0.0, 1.0) * 255.0) as u8,
        (g.clamp(0.0, 1.0) * 255.0) as u8,
        (b.clamp(0.0, 1.0) * 255.0) as u8,
    ]
}

/// Convert thermal Fahrenheit data to an RGBA image using the inferno colormap.
fn thermal_to_rgba(data: &[f64], width: usize, height: usize, vmin: f64, vmax: f64) -> Vec<u8> {
    let range = if (vmax - vmin).abs() < 1e-6 {
        1.0
    } else {
        vmax - vmin
    };
    let mut rgba = Vec::with_capacity(width * height * 4);
    for &val in data.iter().take(width * height) {
        let t = (val - vmin) / range;
        let [r, g, b] = inferno_color(t);
        rgba.push(r);
        rgba.push(g);
        rgba.push(b);
        rgba.push(255);
    }
    rgba
}

struct HeatSeekerApp {
    irg_data: Option<IrgData>,
    thermal_fahrenheit: Vec<f64>,
    texture: Option<egui::TextureHandle>,
    range_min: f64,
    range_max: f64,
    data_min: f64,
    data_max: f64,
    hover_temp: Option<f64>,
    image_width: usize,
    image_height: usize,
    needs_update: bool,
}

impl Default for HeatSeekerApp {
    fn default() -> Self {
        Self {
            irg_data: None,
            thermal_fahrenheit: Vec::new(),
            texture: None,
            range_min: 0.0,
            range_max: 100.0,
            data_min: 0.0,
            data_max: 100.0,
            hover_temp: None,
            image_width: 0,
            image_height: 0,
            needs_update: false,
        }
    }
}

impl HeatSeekerApp {
    fn load_irg(&mut self, path: &str) {
        match irg::extract_irg(path) {
            Ok(data) => {
                let fahrenheit = data.thermal_fahrenheit();
                self.image_width = data.thermal_width;
                self.image_height = data.thermal_height;

                let min_val = fahrenheit
                    .iter()
                    .copied()
                    .fold(f64::INFINITY, f64::min);
                let max_val = fahrenheit
                    .iter()
                    .copied()
                    .fold(f64::NEG_INFINITY, f64::max);

                self.data_min = min_val;
                self.data_max = max_val;
                self.range_min = min_val;
                self.range_max = max_val;
                self.thermal_fahrenheit = fahrenheit;
                self.irg_data = Some(data);
                self.texture = None; // force regeneration
                self.needs_update = true;
            }
            Err(e) => {
                eprintln!("Error loading IRG file: {e}");
            }
        }
    }

    fn update_texture(&mut self, ctx: &egui::Context) {
        if self.thermal_fahrenheit.is_empty() {
            return;
        }
        let rgba = thermal_to_rgba(
            &self.thermal_fahrenheit,
            self.image_width,
            self.image_height,
            self.range_min,
            self.range_max,
        );
        let color_image = egui::ColorImage::from_rgba_unmultiplied(
            [self.image_width, self.image_height],
            &rgba,
        );
        self.texture = Some(ctx.load_texture(
            "thermal",
            color_image,
            egui::TextureOptions::NEAREST,
        ));
        self.needs_update = false;
    }
}

impl eframe::App for HeatSeekerApp {
    fn update(&mut self, ctx: &egui::Context, _frame: &mut eframe::Frame) {
        // Menu bar
        egui::TopBottomPanel::top("menu_bar").show(ctx, |ui| {
            egui::menu::bar(ui, |ui| {
                ui.menu_button("File", |ui| {
                    if ui.button("Open IRG…").clicked() {
                        ui.close_menu();
                        if let Some(path) = rfd::FileDialog::new()
                            .add_filter("IRG files", &["irg"])
                            .pick_file()
                        {
                            self.load_irg(&path.display().to_string());
                        }
                    }
                });
            });
        });

        // Side panel with range controls
        egui::SidePanel::right("controls")
            .min_width(180.0)
            .show(ctx, |ui| {
                ui.heading("Temperature Range (°F)");
                ui.add_space(10.0);

                ui.label("Min:");
                let min_slider = ui.add(
                    egui::Slider::new(&mut self.range_min, self.data_min..=self.data_max)
                        .text("°F"),
                );
                if min_slider.changed() {
                    if self.range_min > self.range_max {
                        self.range_min = self.range_max;
                    }
                    self.needs_update = true;
                }

                ui.label("Max:");
                let max_slider = ui.add(
                    egui::Slider::new(&mut self.range_max, self.data_min..=self.data_max)
                        .text("°F"),
                );
                if max_slider.changed() {
                    if self.range_max < self.range_min {
                        self.range_max = self.range_min;
                    }
                    self.needs_update = true;
                }

                ui.add_space(20.0);
                if ui.button("Reset Range").clicked() {
                    self.range_min = self.data_min;
                    self.range_max = self.data_max;
                    self.needs_update = true;
                }

                ui.add_space(20.0);
                ui.separator();

                // Display hover temperature
                if let Some(temp) = self.hover_temp {
                    ui.heading(format!("{temp:.2} °F"));
                } else {
                    ui.heading("Hover over image");
                }

                // Display header info if loaded
                if let Some(ref data) = self.irg_data {
                    ui.add_space(10.0);
                    ui.separator();
                    ui.label("IRG Info:");
                    ui.label(format!(
                        "Emissivity: {:.4}",
                        data.header.emissivity as f64 / 10000.0
                    ));
                    ui.label(format!(
                        "Ambient: {:.1} K",
                        data.header.ambient_temperature as f64 / 10000.0
                    ));
                    ui.label(format!(
                        "Distance: {:.1} m",
                        data.header.distance as f64 / 10000.0
                    ));
                }
            });

        // Central panel with thermal image
        egui::CentralPanel::default().show(ctx, |ui| {
            if self.needs_update {
                self.update_texture(ctx);
            }

            if let Some(ref texture) = self.texture {
                let available = ui.available_size();
                let aspect = self.image_width as f32 / self.image_height as f32;
                let display_width = available.x.min(available.y * aspect);
                let display_height = display_width / aspect;
                let size = egui::vec2(display_width, display_height);

                let (response, painter) =
                    ui.allocate_painter(size, egui::Sense::hover());

                painter.image(
                    texture.id(),
                    response.rect,
                    egui::Rect::from_min_max(egui::pos2(0.0, 0.0), egui::pos2(1.0, 1.0)),
                    egui::Color32::WHITE,
                );

                // Update hover temperature
                if let Some(pos) = response.hover_pos() {
                    let rel_x =
                        (pos.x - response.rect.min.x) / response.rect.width();
                    let rel_y =
                        (pos.y - response.rect.min.y) / response.rect.height();
                    let col = (rel_x * self.image_width as f32) as usize;
                    let row = (rel_y * self.image_height as f32) as usize;
                    if col < self.image_width && row < self.image_height {
                        let idx = row * self.image_width + col;
                        if idx < self.thermal_fahrenheit.len() {
                            self.hover_temp = Some(self.thermal_fahrenheit[idx]);
                        }
                    }
                } else {
                    self.hover_temp = None;
                }
            } else {
                ui.centered_and_justified(|ui| {
                    ui.heading("Open an IRG file from the File menu");
                });
            }
        });
    }
}

fn main() -> eframe::Result {
    let options = eframe::NativeOptions {
        viewport: egui::ViewportBuilder::default()
            .with_inner_size([900.0, 700.0])
            .with_min_inner_size([400.0, 300.0]),
        ..Default::default()
    };
    eframe::run_native(
        "HeatSeeker",
        options,
        Box::new(|_cc| Ok(Box::new(HeatSeekerApp::default()))),
    )
}
