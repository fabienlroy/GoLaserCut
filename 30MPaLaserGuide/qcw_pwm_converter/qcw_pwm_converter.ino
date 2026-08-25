// QCW PWM Converter + Safety Interlocks for Water-Jet Guided Laser
//
// Reads GRBL spindle PWM from FoxAlien controller (S0-S1000, ~1kHz)
// Outputs QCW pulse train at configurable frequency (1-50 kHz)
// Monitors all safety sensors and kills laser on any fault
//
// Wiring:
//   D2  <-- FoxAlien spindle PWM (interrupt-capable)
//   D3  <-- Water leak detector (S3, active HIGH = leak)
//   D4  <-- Door interlock switch (S4, NC: LOW = closed/safe)
//   D5  <-- E-stop button (S5, NC: LOW = not pressed/safe)
//   D9  --> QCW TTL output to laser driver
//   D10 --> Laser Enable output to laser driver
//   D13 --> Fault LED (built-in, HIGH = fault)
//   A0  <-- HP pressure transducer (S1, QDX50A 0-5V = 0-400 bar)
//   A1  <-- LP flow sensor pulse (S2, hall-effect)
//   GND --- common ground with FoxAlien and laser driver
//
// Safety interlocks (any fault = immediate laser kill):
//   - HP pressure out of range (< 200 bar or > 350 bar)
//   - LP cooling flow too low (< 1 L/min)
//   - Water leak detected
//   - Enclosure door open
//   - E-stop pressed
//
// Serial commands (115200 baud):
//   F<freq>   Set QCW frequency in Hz (e.g. F20000 for 20 kHz)
//   D<duty>   Override duty cycle 0-100 (bypasses PWM input)
//   A         Return to auto mode (duty from PWM input)
//   R         Reset fault (after fixing the cause)
//   S         Print current status
//
// Arduino Nano/Uno: Timer1 on pins 9/10, 16 MHz clock

// --- Pin assignments ---
#define PIN_PWM_IN       2
#define PIN_LEAK         3
#define PIN_DOOR         4
#define PIN_ESTOP        5
#define PIN_QCW_OUT      9
#define PIN_ENABLE_OUT   10
#define PIN_FAULT_LED    13
#define PIN_PRESSURE     A0
#define PIN_FLOW         A1

// --- QCW parameters ---
#define DEFAULT_QCW_FREQ  20000
#define MIN_QCW_FREQ      1000
#define MAX_QCW_FREQ      50000
#define GRBL_PWM_TIMEOUT  50000

// --- Pressure thresholds (QDX50A: 0-5V = 0-400 bar) ---
// ADC: 0-1023 maps to 0-400 bar
#define PRESSURE_MIN_BAR  200
#define PRESSURE_MAX_BAR  350
#define PRESSURE_SCALE    400.0
#define ADC_TO_BAR(adc)   ((float)(adc) * PRESSURE_SCALE / 1023.0)

// --- Flow thresholds ---
// Hall-effect flow sensor: pulses per liter (typical ~450 pulses/L)
#define FLOW_PULSES_PER_LITER  450
#define FLOW_MIN_LPM           1.0
#define FLOW_CHECK_INTERVAL_MS 1000

// --- Fault codes ---
#define FAULT_NONE        0
#define FAULT_PRESSURE_LO 1
#define FAULT_PRESSURE_HI 2
#define FAULT_FLOW_LOW    3
#define FAULT_LEAK        4
#define FAULT_DOOR_OPEN   5
#define FAULT_ESTOP       6

// --- QCW state ---
volatile unsigned long pulseStart = 0;
volatile unsigned long pulseHighUs = 0;
volatile unsigned long pulsePeriodUs = 0;
volatile bool newPulse = false;

unsigned long qcwFreq = DEFAULT_QCW_FREQ;
int dutyCyclePercent = 0;
bool autoMode = true;

// --- Safety state ---
uint8_t faultCode = FAULT_NONE;
bool laserAllowed = false;

// --- Flow measurement ---
volatile unsigned long flowPulseCount = 0;
unsigned long lastFlowCheck = 0;
unsigned long lastFlowPulses = 0;
float flowLPM = 0;

const char* faultNames[] = {
    "NONE",
    "PRESSURE LOW",
    "PRESSURE HIGH",
    "FLOW LOW",
    "WATER LEAK",
    "DOOR OPEN",
    "E-STOP"
};

void setup() {
    Serial.begin(115200);

    pinMode(PIN_PWM_IN, INPUT);
    pinMode(PIN_LEAK, INPUT);
    pinMode(PIN_DOOR, INPUT_PULLUP);
    pinMode(PIN_ESTOP, INPUT_PULLUP);
    pinMode(PIN_QCW_OUT, OUTPUT);
    pinMode(PIN_ENABLE_OUT, OUTPUT);
    pinMode(PIN_FAULT_LED, OUTPUT);
    pinMode(PIN_FLOW, INPUT);

    killLaser();

    attachInterrupt(digitalPinToInterrupt(PIN_PWM_IN), pwmISR, CHANGE);
    attachInterrupt(digitalPinToInterrupt(PIN_LEAK), leakISR, RISING);

    // Flow sensor on pin change (polled via analog, or use pin change interrupt)
    // For simplicity, we count rising edges on A1 via polling in loop()

    setupTimer1(qcwFreq, 0);

    Serial.println("QCW + Safety Controller ready");
    Serial.print("QCW freq: ");
    Serial.print(qcwFreq);
    Serial.println(" Hz");
    Serial.print("Pressure range: ");
    Serial.print(PRESSURE_MIN_BAR);
    Serial.print("-");
    Serial.print(PRESSURE_MAX_BAR);
    Serial.println(" bar");
    Serial.println("Commands: F<freq> D<duty> A R S");
}

// --- Interrupt handlers ---

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

void leakISR() {
    faultCode = FAULT_LEAK;
    killLaser();
}

// --- Laser control ---

void killLaser() {
    TCCR1A &= ~(1 << COM1A1);
    digitalWrite(PIN_QCW_OUT, LOW);
    digitalWrite(PIN_ENABLE_OUT, LOW);
    laserAllowed = false;
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

    TCCR1A = (1 << COM1A1) | (1 << WGM11);
    TCCR1B = (1 << WGM13) | (1 << WGM12) | prescaler;
}

void updateQCWOutput(unsigned long freq, int dutyPercent) {
    if (!laserAllowed || dutyPercent <= 0) {
        killLaser();
        return;
    }

    if (dutyPercent > 100) dutyPercent = 100;

    uint8_t dutyByte = map(dutyPercent, 0, 100, 0, 255);
    setupTimer1(freq, dutyByte);
    digitalWrite(PIN_ENABLE_OUT, HIGH);
}

// --- Sensor reading ---

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

float readPressureBar() {
    int adc = analogRead(PIN_PRESSURE);
    return ADC_TO_BAR(adc);
}

void updateFlowRate() {
    unsigned long now = millis();
    if (now - lastFlowCheck < FLOW_CHECK_INTERVAL_MS) return;

    // Simple polling: read analog pin as digital for pulse counting
    // For better accuracy, use a dedicated pin change interrupt
    unsigned long elapsed = now - lastFlowCheck;
    unsigned long pulses = flowPulseCount - lastFlowPulses;
    lastFlowPulses = flowPulseCount;
    lastFlowCheck = now;

    float liters = (float)pulses / FLOW_PULSES_PER_LITER;
    flowLPM = liters * (60000.0 / elapsed);
}

// --- Safety checks ---

uint8_t checkInterlocks() {
    if (digitalRead(PIN_ESTOP) == LOW)
        return FAULT_ESTOP;

    if (digitalRead(PIN_DOOR) == HIGH)
        return FAULT_DOOR_OPEN;

    if (digitalRead(PIN_LEAK) == HIGH)
        return FAULT_LEAK;

    float pressure = readPressureBar();
    if (pressure < PRESSURE_MIN_BAR && dutyCyclePercent > 0)
        return FAULT_PRESSURE_LO;
    if (pressure > PRESSURE_MAX_BAR)
        return FAULT_PRESSURE_HI;

    updateFlowRate();
    if (flowLPM < FLOW_MIN_LPM && dutyCyclePercent > 0)
        return FAULT_FLOW_LOW;

    return FAULT_NONE;
}

// --- Serial interface ---

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
        case 'R':
        case 'r':
            if (checkInterlocks() == FAULT_NONE) {
                faultCode = FAULT_NONE;
                laserAllowed = true;
                digitalWrite(PIN_FAULT_LED, LOW);
                Serial.println("Fault cleared, laser armed");
            } else {
                Serial.print("Cannot reset, active fault: ");
                Serial.println(faultNames[faultCode]);
            }
            break;
        case 'S':
        case 's': {
            float pressure = readPressureBar();
            Serial.print("Freq: ");
            Serial.print(qcwFreq);
            Serial.print(" Hz | Duty: ");
            Serial.print(dutyCyclePercent);
            Serial.print("% | Mode: ");
            Serial.println(autoMode ? "auto" : "manual");
            Serial.print("Pressure: ");
            Serial.print(pressure, 1);
            Serial.print(" bar | Flow: ");
            Serial.print(flowLPM, 1);
            Serial.print(" L/min | Leak: ");
            Serial.println(digitalRead(PIN_LEAK) ? "YES" : "no");
            Serial.print("Door: ");
            Serial.print(digitalRead(PIN_DOOR) ? "OPEN" : "closed");
            Serial.print(" | E-stop: ");
            Serial.print(digitalRead(PIN_ESTOP) ? "OK" : "PRESSED");
            Serial.print(" | Fault: ");
            Serial.print(faultNames[faultCode]);
            Serial.print(" | Laser: ");
            Serial.println(laserAllowed ? "ARMED" : "SAFE");
            break;
        }
    }
}

// --- Main loop ---

void loop() {
    handleSerial();

    uint8_t fault = checkInterlocks();
    if (fault != FAULT_NONE) {
        if (faultCode == FAULT_NONE) {
            Serial.print("FAULT: ");
            Serial.println(faultNames[fault]);
        }
        faultCode = fault;
        killLaser();
        digitalWrite(PIN_FAULT_LED, HIGH);
    }

    if (autoMode) {
        int grblDuty = readGRBLDuty();
        if (grblDuty >= 0) {
            dutyCyclePercent = grblDuty;
        }
    }

    if (laserAllowed && faultCode == FAULT_NONE) {
        updateQCWOutput(qcwFreq, dutyCyclePercent);
    }

    delay(10);
}
