import { useState, useEffect } from "react";
import axios from "axios";
import { MdOutlineWatchLater } from "react-icons/md";

function App() {
  // Helper function to convert text to sentence case.
  const sentenceCase = (text) => {
    if (!text) return "";
    return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
  };

  // State for courses list and form inputs.
  const [courses, setCourses] = useState([]);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    duration: "",
    image: null,
    pdf: null,
  });

  // Fetch courses from the backend.
  const fetchCourses = () => {
    axios
      .get("http://localhost:8080/courses")
      .then((response) => {
        setCourses(response.data);
      })
      .catch((error) => {
        console.error("Error fetching courses:", error);
      });
  };

  useEffect(() => {
    fetchCourses();
  }, []);

  // Handle input changes for text and file inputs.
  const handleChange = (e) => {
    if (e.target.name === "image" || e.target.name === "pdf") {
      setFormData({ ...formData, [e.target.name]: e.target.files[0] });
    } else {
      setFormData({ ...formData, [e.target.name]: e.target.value });
    }
  };

  // Handle form submission for adding a course.
  const handleSubmit = (e) => {
    e.preventDefault();

    // Client-side validation for description length.
    if (formData.description.length > 100) {
      alert("Description must be 100 characters or less");
      return;
    }

    const data = new FormData();
    data.append("title", formData.title);
    data.append("description", formData.description);
    data.append("duration", formData.duration);
    data.append("image", formData.image);
    data.append("pdf", formData.pdf);

    axios
      .post("http://localhost:8080/add-course", data, {
        headers: { "Content-Type": "multipart/form-data" },
      })
      .then((response) => {
        alert(response.data.message);
        setFormData({
          title: "",
          description: "",
          duration: "",
          image: null,
          pdf: null,
        });
        fetchCourses();
      })
      .catch((error) => {
        alert("Failed to add course");
        console.error(error);
      });
  };

  return (
    <div style={{ padding: "20px" }}>
      <h1>Add Course</h1>
      <form onSubmit={handleSubmit} style={{ marginBottom: "40px" }}>
        <input
          type="text"
          name="title"
          placeholder="Title"
          value={formData.title}
          onChange={handleChange}
          required
          style={{ display: "block", margin: "10px 0", width: "300px" }}
        />
        <textarea
          name="description"
          placeholder="Description (max 100 characters)"
          value={formData.description}
          onChange={handleChange}
          required
          maxLength={100}
          style={{
            display: "block",
            margin: "10px 0",
            width: "300px",
            height: "100px",
          }}
        />
        <input
          type="number"
          name="duration"
          placeholder="Duration"
          value={formData.duration}
          onChange={handleChange}
          required
          style={{ display: "block", margin: "10px 0", width: "300px" }}
        />
        <div style={{ margin: "10px 0" }}>
          <label>
            Choose Image:
            <input
              type="file"
              name="image"
              onChange={handleChange}
              required
              accept="image/png, image/jpeg, image/jpg"
              style={{ marginLeft: "10px" }}
            />
          </label>
        </div>
        <div style={{ margin: "10px 0" }}>
          <label>
            Choose PDF:
            <input
              type="file"
              name="pdf"
              onChange={handleChange}
              required
              accept="application/pdf"
              style={{ marginLeft: "10px" }}
            />
          </label>
        </div>
        <button type="submit" style={{ padding: "10px 20px" }}>
          Submit
        </button>
      </form>

      <h1>Courses</h1>
      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "20px",
        }}
      >
        {courses.map((course) => (
          <div
            key={course.id}
            style={{
              border: "1px solid #ccc",
              borderRadius: "8px",
              padding: "16px",
              width: "300px",
              boxShadow: "0 2px 5px rgba(0,0,0,0.1)",
              wordWrap: "break-word",
              whiteSpace: "normal",
              display: "flex",
              flexDirection: "column",
              justifyContent: "space-between",
              minHeight: "180px",
            }}
          >
            {course.image && (
              <img
                src={course.image}
                alt={course.title}
                style={{
                  width: "100%",
                  height: "200px",
                  objectFit: "cover",
                  borderRadius: "8px",
                }}
              />
            )}
            <h3 style={{ textTransform: "uppercase", fontWeight: "bold" }}>
              {course.title}
            </h3>
            <p>{sentenceCase(course.description)}</p>
            {/* Bottom Section: Watch icon, duration and Start Learning link */}
            <div
              style={{
                marginTop: "auto",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
              }}
            >
              <p style={{ display: "flex", alignItems: "center", margin: 0 }}>
                <MdOutlineWatchLater style={{ marginRight: "6px" }} />
                {course.duration}
              </p>
              <a
                href="https://example.com/start-learning" // Replace with your actual URL.
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  textDecoration: "none",
                  color: "#007bff",
                  fontWeight: "bold",
                }}
              >
                Start Learning
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
