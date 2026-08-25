// QCW PWM Converter for Water-Jet Guided Laser
//
// Reads GRBL spindle PWM from FoxAlien controller (S0-S1000, ~1kHz)
// Outputs QCW pulse train at configurable frequency (1-50 kHz)
// S-value controls duty cycle, QCW frequency set via serial or potentiometer
//
// Wiring:
//   FoxAlien spindle PWM  --> pin 2 (input, interrupt-capable)
//   QCW TTL output        --> pin 9 (Timer1 PWM output)
//   Laser Enable output   --> pin 10 (HIGH when S > 0)
//   Frequency pot (opt.)  --> A0 (10k potentiometer, center wiper)
//   GND                   --> common ground with FoxAlien and laser driver
//
// Serial commands (115200 baud):
//   F<freq>   Set QCW frequency in Hz (e.g. F20000 for 20 kHz)
//   D<duty>   Override duty cycle 0-100 (bypasses PWM input)
//   A         Return to auto mode (duty from PWM input)
//   S         Print current status
//
// Arduino Nano/Uno: Timer1 on pins 9/10, 16 MHz clock

#define PIN_PWM_IN      2
#define PIN_QCW_OUT     9
#define PIN_ENABLE_OUT  10
#define PIN_FREQ_POT    A0

#define DEFAULT_QCW_FREQ  20000
#define MIN_QCW_FREQ      1000
#define MAX_QCW_FREQ      50000
#define GRBL_PWM_TIMEOUT  50000

volatile unsigned long pulseStart = 0;
volatile unsigned long pulseHighUs = 0;
volatile unsigned long pulsePeriodUs = 0;
volatile bool newPulse = false;

unsigned long qcwFreq = DEFAULT_QCW_FREQ;
int dutyCyclePercent = 0;
bool autoMode = true;
bool usePot = false;

void setup() {
    Serial.begin(115200);

    pinMode(PIN_PWM_IN, INPUT);
    pinMode(PIN_QCW_OUT, OUTPUT);
    pinMode(PIN_ENABLE_OUT, OUTPUT);

    digitalWrite(PIN_QCW_OUT, LOW);
    digitalWrite(PIN_ENABLE_OUT, LOW);

    attachInterrupt(digitalPinToInterrupt(PIN_PWM_IN), pwmISR, CHANGE);

    setupTimer1(qcwFreq, 0);

    Serial.println("QCW PWM Converter ready");
    Serial.print("Frequency: ");
    Serial.print(qcwFreq);
    Serial.println(" Hz");
    Serial.println("Commands: F<freq> D<duty> A S");
}

void pwmISR() {
    unsigned long now = micros();
    if (digitalRead(PIN_PWM_IN) == HIGH) {
        pulsePeriodUs = now - pulseStart;
        pulseStart = now;
    } else {
        pulseHighUs = now - pulseStart;
        newPulse = true;
    }
}

void setupTimer1(unsigned long freq, uint8_t dutyByte) {
    TCCR1A = 0;
    TCCR1B = 0;

    unsigned long topValue = (16000000UL / freq) - 1;
    uint8_t prescaler;

    if (topValue <= 65535) {
        prescaler = (1 << CS10);
    } else if (topValue / 8 <= 65535) {
        topValue = (16000000UL / (8 * freq)) - 1;
        prescaler = (1 << CS11);
    } else {
        topValue = (16000000UL / (64 * freq)) - 1;
        prescaler = (1 << CS11) | (1 << CS10);
    }

    if (topValue > 65535) topValue = 65535;

    ICR1 = topValue;
    unsigned long ocrValue = ((unsigned long)topValue * dutyByte) / 255;
    OCR1A = ocrValue;

    // Fast PWM, TOP = ICR1
    TCCR1A = (1 << COM1A1) | (1 << WGM11);
    TCCR1B = (1 << WGM13) | (1 << WGM12) | prescaler;
}

void updateQCWOutput(unsigned long freq, int dutyPercent) {
    if (dutyPercent <= 0) {
        TCCR1A &= ~(1 << COM1A1);
        digitalWrite(PIN_QCW_OUT, LOW);
        digitalWrite(PIN_ENABLE_OUT, LOW);
        return;
    }

    if (dutyPercent > 100) dutyPercent = 100;

    uint8_t dutyByte = map(dutyPercent, 0, 100, 0, 255);
    setupTimer1(freq, dutyByte);
    digitalWrite(PIN_ENABLE_OUT, HIGH);
}

int readGRBLDuty() {
    noInterrupts();
    unsigned long highUs = pulseHighUs;
    unsigned long periodUs = pulsePeriodUs;
    bool valid = newPulse;
    newPulse = false;
    interrupts();

    if (!valid || periodUs == 0) return -1;
    if (periodUs > GRBL_PWM_TIMEOUT) return 0;

    int duty = constrain((int)((highUs * 100) / periodUs), 0, 100);
    return duty;
}

void handleSerial() {
    if (!Serial.available()) return;

    char cmd = Serial.read();
    switch (cmd) {
        case 'F':
        case 'f': {
            unsigned long f = Serial.parseInt();
            if (f >= MIN_QCW_FREQ && f <= MAX_QCW_FREQ) {
                qcwFreq = f;
                Serial.print("QCW freq: ");
                Serial.print(qcwFreq);
                Serial.println(" Hz");
            } else {
                Serial.println("ERR: freq 1000-50000");
            }
            break;
        }
        case 'D':
        case 'd': {
            int d = Serial.parseInt();
            if (d >= 0 && d <= 100) {
                autoMode = false;
                dutyCyclePercent = d;
                Serial.print("Manual duty: ");
                Serial.print(d);
                Serial.println("%");
            } else {
                Serial.println("ERR: duty 0-100");
            }
            break;
        }
        case 'A':
        case 'a':
            autoMode = true;
            Serial.println("Auto mode (duty from PWM input)");
            break;
        case 'S':
        case 's':
            Serial.print("Freq: ");
            Serial.print(qcwFreq);
            Serial.print(" Hz | Duty: ");
            Serial.print(dutyCyclePercent);
            Serial.print("% | Mode: ");
            Serial.print(autoMode ? "auto" : "manual");
            Serial.print(" | Enable: ");
            Serial.println(digitalRead(PIN_ENABLE_OUT) ? "ON" : "OFF");
            break;
    }
}

void loop() {
    handleSerial();

    if (autoMode) {
        int grblDuty = readGRBLDuty();
        if (grblDuty >= 0) {
            dutyCyclePercent = grblDuty;
        }
    }

    updateQCWOutput(qcwFreq, dutyCyclePercent);
    delay(10);
}
